package exporter

import (
	"fmt"
	"os"

	"github.com/docker/docker/daemon/logger"
	"github.com/moby/moby/daemon/logger/awslogs"
	"github.com/previousnext/k8s-aws-cloudwatchlogs/k8slog"
)

// Push is a used setup an AWS CloudWatchLogs Group/Stream. We then consume a stream and push.
func Push(region, dir string, file os.FileInfo, new bool) error {
	kl, err := k8slog.New(dir, file)
	if err != nil {
		return err
	}

	// We load the stream backend so we can push these logs to the channel.
	cw, err := awslogs.New(logger.Info{
		Config: map[string]string{
			"awslogs-region":       region,
			"awslogs-create-group": "true",
			"awslogs-group":        kl.Namespace,
			"awslogs-stream":       fmt.Sprintf("%s-%s", kl.Pod, kl.Container),
		},
	})
	if err != nil {
		return err
	}
	defer cw.Close()

	stream, err := kl.Stream(new)
	if err != nil {
		return err
	}

	for message := range stream {
		err = cw.Log(&logger.Message{
			Line:      []byte(message.Log),
			Timestamp: message.Time,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
