package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/previousnext/k8s-aws-cloudwatchlogs/exporter"
	"github.com/previousnext/k8s-aws-cloudwatchlogs/source"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	cliDirectory = kingpin.Flag("dir", "The Kubernetes container logs directory").Default("/var/log/containers").OverrideDefaultFromEnvar("KUBERNETES_CONTAINER_LOGS").String()
	cliRegion    = kingpin.Flag("region", "The AWS region to store the logs").Default("ap-southeast-2").OverrideDefaultFromEnvar("AWS_REGION").String()
)

func main() {
	kingpin.Parse()

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
		go push(file, false)
	}

	for {
		file := <-created
		go push(file, true)
	}
}

func push(file os.FileInfo, new bool) {
	err := exporter.Push(*cliRegion, *cliDirectory, file, new)
	if err != nil {
		log.WithFields(log.Fields{"file": file.Name()}).Error("Failed to push file: %s", err)
	} else {
		log.WithFields(log.Fields{"file": file.Name()}).Info("Finished pushing file")
	}
}
