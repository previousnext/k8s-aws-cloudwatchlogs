package main

import (
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	cliEndpoint = kingpin.Flag("docker", "The Docker endpoint").Default("unix:///var/run/docker.sock").OverrideDefaultFromEnvar("DOCKER").String()
	cliRegion   = kingpin.Flag("region", "The AWS region to store the logs").Default("ap-southeast-2").OverrideDefaultFromEnvar("REGION").String()
)

func main() {
	kingpin.Parse()

	// Determine the hostname (name of the pod) so we can use it for filtering.
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	log.Info("Connecting to Docker...")

	client, err := docker.NewClient(*cliEndpoint)
	if err != nil {
		panic(err)
	}

	containers, err := client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		panic(err)
	}

	// Load the existing containers that are running on this host.
	for _, container := range containers {
		namespace, pod, name, err := filter(container.Labels, hostname)
		if err != nil {
			log.WithFields(log.Fields{"container": container.ID}).Info(err)
			continue
		}

		go func(c docker.APIContainers) {
			err = stream(client, *cliRegion, container.ID, namespace, fmt.Sprintf("%s-%s", pod, name))
			if err != nil {
				log.WithFields(log.Fields{"container": container.ID}).Info(err)
			}
		}(container)
	}

	// Listening for changes in container state.
	//  * New containers
	//  * Removing old containers
	listener := make(chan *docker.APIEvents)
	err = client.AddEventListener(listener)
	if err != nil {
		log.Fatal(err)
	}

	// This will ensure our events listener will close.
	defer func() {
		err = client.RemoveEventListener(listener)
		if err != nil {
			log.Fatal(err)
		}
	}()

	limiter := time.Tick(time.Second * 10)
	for {
		select {
		case msg := <-listener:
			// This means we can register another routine for listening to
			// stdout and stderr of these containers.
			if msg.Status == "start" {
				container, err := client.InspectContainer(msg.ID)
				if err != nil {
					log.WithFields(log.Fields{"container": msg.ID}).Info(err)
					continue
				}

				namespace, pod, name, err := filter(container.Config.Labels, hostname)
				if err != nil {
					log.WithFields(log.Fields{"container": container.ID}).Info(err)
					continue
				}

				go stream(client, *cliRegion, container.ID, namespace, fmt.Sprintf("%s-%s", pod, name))
			}
		case <-limiter:
			break
		}
	}
}
