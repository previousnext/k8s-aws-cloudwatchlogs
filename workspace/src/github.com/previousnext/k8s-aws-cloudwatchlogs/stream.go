package main

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/daemon/logger"
	"github.com/fsouza/go-dockerclient"
	"github.com/moby/moby/daemon/logger/awslogs"
)

func stream(client *docker.Client, region, container, group, stream string) error {
	log.WithFields(log.Fields{"group": group, "stream": stream, "container": container}).Info("Started capturing logs")

	var (
		stdoutBuffer bytes.Buffer
		stderrBuffer bytes.Buffer
	)

	exit := make(chan bool)

	go func() {
		client.Logs(docker.LogsOptions{
			Since:        time.Now().Unix(),
			Container:    container,
			Follow:       true,
			Stdout:       true,
			Stderr:       true,
			Tail:         "all",
			Timestamps:   true,
			RawTerminal:  false,
			OutputStream: &stdoutBuffer,
			ErrorStream:  &stderrBuffer,
		})
		close(exit)
	}()

	stdoutCh := readerToChannel(&stdoutBuffer, exit)
	stderrCh := readerToChannel(&stderrBuffer, exit)

	// We load the stream backend so we can push these logs to the channel.
	cw, err := awslogs.New(logger.Info{
		Config: map[string]string{
			"awslogs-region":       region,
			"awslogs-create-group": group,
			"awslogs-stream":       stream,
		},
	})
	if err != nil {
		return err
	}

	// Listen for stdout or stderr messages from the channels. We then hand these
	// off to the queue backend.
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for value := range stdoutCh {
			err := cw.Log(&logger.Message{
				Line: []byte(value),
			})
			if err != nil {
				log.WithFields(log.Fields{"group": group, "stream": stream, "container": container}).Info("Stdout: %s", err)
			}
		}
	}()
	go func() {
		defer wg.Done()
		for value := range stderrCh {
			err := cw.Log(&logger.Message{
				Line: []byte(value),
			})
			if err != nil {
				log.WithFields(log.Fields{"group": group, "stream": stream, "container": container}).Info("Stderr: %s", err)
			}
		}
	}()
	wg.Wait()

	log.WithFields(log.Fields{"group": group, "stream": stream, "container": container}).Info("Finished capturing logs")

	return nil
}

// Helper function to handle log streams and massage them into a string representation.
func readerToChannel(reader *bytes.Buffer, exit <-chan bool) <-chan string {
	channel := make(chan string)

	limiter := time.Tick(100 * time.Millisecond)

	go func() {
		for {
			select {
			case <-exit:
				close(channel)
				return
			default:
				<-limiter

				// This avoids issues when trying to ReadSring and there is no data
				// being passed in.
				if reader.Len() <= 0 {
					continue
				}

				line, err := reader.ReadString('\n')
				if err != nil && err != io.EOF {
					close(channel)
					return
				}

				line = strings.TrimSpace(line)
				if line != "" {
					channel <- line
				}
			}
		}
	}()

	return channel
}
