package main

import (
	"bufio"
	"fmt"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/daemon/logger"
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

	fmt.Println("Logs")

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

	r := bufio.NewReader(rc)

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return err
		}

		line = strings.TrimSpace(line)
		if line != "" {
			err := cw.Log(&logger.Message{
				Line:      []byte(line),
				Timestamp: time.Now(),
			})
			if err != nil {
				log.WithFields(log.Fields{"group": group, "stream": stream, "container": container}).Info("Stdout: %s", err)
			}
		}
	}

	log.WithFields(log.Fields{"group": group, "stream": stream, "container": container}).Info("Finished capturing logs")

	return nil
}
