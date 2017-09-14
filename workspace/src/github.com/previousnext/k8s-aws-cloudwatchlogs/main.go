package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/daemon/logger"
	"github.com/hpcloud/tail"
	"github.com/moby/moby/daemon/logger/awslogs"
	"github.com/previousnext/k8s-aws-cloudwatchlogs/source"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	cliDirectory  = kingpin.Flag("dir", "The Kubernetes container logs directory").Default("/var/log/containers").OverrideDefaultFromEnvar("KUBERNETES_CONTAINER_LOGS").String()
	cliRegion     = kingpin.Flag("region", "The AWS region to store the logs").Default("ap-southeast-2").OverrideDefaultFromEnvar("AWS_REGION").String()
	cliPrometheus = kingpin.Flag("prometheus", "Prometheus metrics endpoint").Default(":9000").OverrideDefaultFromEnvar("PROMETHEUS").String()
)

func main() {
	kingpin.Parse()

	log.Info("Serving Prometheus metrics endpoint")

	go metrics(*cliPrometheus)

	log.Info("Retrieving a list of existing log files")

	existing, err := source.List(*cliDirectory)
	if err != nil {
		panic(err)
	}

	log.Info("Watching directory for new log files")

	created, err := source.Watch(*cliDirectory)
	if err != nil {
		panic(err)
	}

	log.Info("Starting to push to remote storage")

	for _, file := range existing {
		go func(file os.FileInfo) {
			log.Infof("Starting to push existing file: %s", file.Name())

			err := push(file, false)
			if err != nil {
				log.Errorf("Failed to push existing file: %s: %s", file.Name(), err)
			} else {
				log.Infof("Finished pushing existing file: %s", file.Name())
			}
		}(file)
	}

	for {
		file, more := <-created
		if !more {
			break
		}

		go func(file os.FileInfo) {
			log.Infof("Starting to push new file: %s", file.Name())

			err := push(file, true)
			if err != nil {
				log.Errorf("Failed to push new file: %s: %s", file.Name(), err)
			} else {
				log.Infof("Finished pushing new file: %s", file.Name())
			}
		}(file)
	}
}

func push(file os.FileInfo, new bool) error {
	namespace, pod, container, err := fileMetadata(file)
	if err != nil {
		return fmt.Errorf("failed to extract namespace, pod and container metadata: %s", err)
	}

	// We load the stream backend so we can push these logs to the channel.
	cw, err := awslogs.New(logger.Info{
		Config: map[string]string{
			"awslogs-region":       *cliRegion,
			"awslogs-create-group": "true",
			"awslogs-group":        namespace,
			"awslogs-stream":       fmt.Sprintf("%s-%s", pod, container),
		},
	})
	if err != nil {
		return err
	}
	defer cw.Close()

	watch, err := fileTail(filepath.Join(*cliDirectory, file.Name()), new)
	if err != nil {
		return fmt.Errorf("failed to start tail: %s", err)
	}

	// We also want to monitor to make sure that the file still exists.
	go func(watch *tail.Tail) {
		limiter := time.Tick(time.Second * 15)

		for {
			<-limiter

			if _, err := os.Stat(watch.Filename); os.IsNotExist(err) {
				watch.Stop()
				return
			}
		}
	}(watch)

	for {
		line, more := <-watch.Lines
		if !more {
			break
		}

		var message Message

		err := json.Unmarshal([]byte(line.Text), &message)
		if err != nil {
			return fmt.Errorf("failed to unmarshal line for %s/%s/%s: %s", namespace, pod, container, err)
		}

		err = cw.Log(&logger.Message{
			Line:      []byte(message.Log),
			Timestamp: message.Time,
		})
		if err != nil {
			return fmt.Errorf("failed to push log: %s", err)
		}
	}

	return nil
}

func metrics(port string) {
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(port, nil))
}
