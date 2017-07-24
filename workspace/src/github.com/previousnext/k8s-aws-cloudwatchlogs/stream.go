package main

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/daemon/logger"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/moby/moby/daemon/logger/awslogs"
	"github.com/samalba/dockerclient"
)

func stream(client *dockerclient.DockerClient, region, container, group, stream string) error {
	log.WithFields(log.Fields{"group": group, "stream": stream, "container": container}).Info("Started capturing logs")

	rc, err := client.ContainerLogs(container, &dockerclient.LogOptions{
		Follow:     true,
		Stdout:     true,
		Stderr:     true,
		Timestamps: false,
		Tail:       1,
	})
	defer rc.Close()

	// We load the stream backend so we can push these logs to the channel.
	cw, err := awslogs.New(logger.Info{
		Config: map[string]string{
			"awslogs-region":       region,
			"awslogs-create-group": "true",
			"awslogs-group":        group,
			"awslogs-stream":       stream,
		},
	})
	if err != nil {
		return err
	}

	var (
		stdout = bytes.NewBuffer(nil)
		stderr = bytes.NewBuffer(nil)
	)

	go func() {
		// Refer to this doc block for why we pass this through stdcopy.
		// https://github.com/moby/moby/blob/master/client/container_logs.go#L23
		_, err = stdcopy.StdCopy(stdout, stderr, rc)
		if err != nil && err != io.EOF {
			log.WithFields(log.Fields{"group": group, "stream": stream, "container": container}).Info(err)
			return
		}
	}()

	var (
		exit     = make(chan bool)
		stdoutCh = readerToChannel(stdout, exit)
		stderrCh = readerToChannel(stderr, exit)
	)

	// Listen for stdout or stderr messages from the channels. We then hand these
	// off to the queue backend.
	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()

		for line := range stdoutCh {
			err := cw.Log(&logger.Message{
				Line:      []byte(line),
				Timestamp: time.Now(),
			})
			if err != nil {
				log.WithFields(log.Fields{"group": group, "stream": stream, "container": container}).Info("Stdout: %s", err)
			}
		}
	}()

	go func() {
		defer wg.Done()

		for line := range stderrCh {
			err := cw.Log(&logger.Message{
				Line:      []byte(line),
				Timestamp: time.Now(),
			})
			if err != nil {
				log.WithFields(log.Fields{"group": group, "stream": stream, "container": container}).Info("Stdout: %s", err)
			}
		}
	}()

	wg.Wait()

	log.WithFields(log.Fields{"group": group, "stream": stream, "container": container}).Info("Finished capturing logs")

	return nil
}

// Helper function to handle log streams and massage them into a string representation.
func readerToChannel(reader *bytes.Buffer, exit <-chan bool) <-chan string {
	var (
		channel = make(chan string)
		limiter = time.Tick(100 * time.Millisecond)
	)

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
