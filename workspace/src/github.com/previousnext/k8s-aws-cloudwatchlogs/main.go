package main

import (
	"fmt"
	"os"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/samalba/dockerclient"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	cliEndpoint = kingpin.Flag("docker", "The Docker endpoint").Default("unix:///var/run/docker.sock").OverrideDefaultFromEnvar("DOCKER").String()
	cliRegion   = kingpin.Flag("region", "The AWS region to store the logs").Default("ap-southeast-2").OverrideDefaultFromEnvar("REGION").String()
)

func main() {
	kingpin.Parse()

	var wg sync.WaitGroup

	// Determine the hostname (name of the pod) so we can use it for filtering.
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	log.Info("Connecting to Docker...")

	client, err := dockerclient.NewDockerClient(*cliEndpoint, nil)
	if err != nil {
		panic(err)
	}

	containers, err := client.ListContainers(false, false, "")
	if err != nil {
		panic(err)
	}

	// Load the existing containers that are running on this host.
	for _, container := range containers {
		namespace, pod, name, err := filter(container.Labels, hostname)
		if err != nil {
			log.WithFields(log.Fields{"container": container.Id}).Info(err)
			continue
		}

		wg.Add(1)

		go func(container dockerclient.Container) {
			defer wg.Done()

			err = stream(client, *cliRegion, container.Id, namespace, fmt.Sprintf("%s-%s", pod, name))
			if err != nil {
				log.WithFields(log.Fields{"container": container.Id}).Info(err)
			}
		}(container)
	}

	// This will allow us to track new containers.
	client.StartMonitorEvents(func(event *dockerclient.Event, ec chan error, args ...interface{}) {
		// Only add our stream listener to new containers.
		if event.Status != "start" {
			return
		}

		container, err := client.InspectContainer(event.ID)
		if err != nil {
			log.WithFields(log.Fields{"container": event.ID}).Info(err)
			return
		}

		namespace, pod, name, err := filter(container.Config.Labels, hostname)
		if err != nil {
			log.WithFields(log.Fields{"container": container.Id}).Info(err)
			return
		}

		wg.Add(1)

		go func(container *dockerclient.ContainerInfo) {
			defer wg.Done()

			err = stream(client, *cliRegion, container.Id, namespace, fmt.Sprintf("%s-%s", pod, name))
			if err != nil {
				log.WithFields(log.Fields{"container": container.Id}).Info(err)
			}
		}(container)
	}, nil)

	wg.Wait()
}
