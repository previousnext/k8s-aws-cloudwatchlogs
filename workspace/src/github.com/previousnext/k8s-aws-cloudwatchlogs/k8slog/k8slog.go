package k8slog

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/gobwas/glob"
	"github.com/hpcloud/tail"
)

const format = "*_*_*-*.log"

// Validate is used to check that our file has the correct naming strategy.
func Validate(file os.FileInfo) bool {
	// Ensure we are not scanning a directory.
	if file.IsDir() {
		return false
	}

	// Make sure that the file we looking for has the correct pattern.
	if !glob.MustCompile(format).Match(file.Name()) {
		return false
	}

	return true
}

// New sets up a new file for streaming.
func New(dir string, file os.FileInfo) (*LogFile, error) {
	namespace, pod, container, err := metadata(file)
	if err != nil {
		return nil, fmt.Errorf("failed to extract namespace, pod and container info: %s", format)
	}

	return &LogFile{
		Namespace: namespace,
		Pod:       pod,
		Container: container,
		path:      fmt.Sprintf("%s/%s", dir, file.Name()),
	}, nil
}

// Stream returns a stream of logs from the file.
func (lf LogFile) Stream(start bool) (chan Log, error) {
	logs := make(chan Log)

	config := tail.Config{
		Follow:    true,
		MustExist: true,
	}

	if !start {
		config.Location = &tail.SeekInfo{
			Offset: 0,
			Whence: io.SeekEnd,
		}
	}

	t, err := tail.TailFile(lf.path, config)
	if err != nil {
		return logs, err
	}

	for line := range t.Lines {
		var l Log

		err := json.Unmarshal([]byte(line.Text), &l)
		if err != nil {
			log.WithFields(log.Fields{
				"namespace": lf.Namespace,
				"pod":       lf.Pod,
				"container": lf.Container,
			}).Fatal(err)
		}

		logs <- l
	}

	return logs, nil
}

func metadata(file os.FileInfo) (string, string, string, error) {
	if !Validate(file) {
		return "", "", "", fmt.Errorf("not a valid file with format: %s", format)
	}

	// Remove the ".log" from the file.
	name := strings.Replace(file.Name(), ".log", "", 0)

	// Split the string down so we can return their metadata.
	var (
		pod       = strings.Split(name, "_")
		container = strings.Split(pod[2], "-")
	)

	return pod[0], pod[1], container[0], nil
}
