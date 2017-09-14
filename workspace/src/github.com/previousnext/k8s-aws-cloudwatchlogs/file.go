package main

import (
	"io"
	"os"
	"strings"

	"fmt"
	"github.com/hpcloud/tail"
	"github.com/previousnext/k8s-aws-cloudwatchlogs/source"
)

func fileMetadata(file os.FileInfo) (string, string, string, error) {
	if !source.Valid(file) {
		return "", "", "", fmt.Errorf("invalid file format: %s", file.Name())
	}

	// Remove the ".log" from the file.
	name := strings.Replace(file.Name(), ".log", "", 1)

	// Split the string down so we can return their metadata.
	sl := strings.Split(name, "_")

	return sl[1], sl[0], sl[2], nil
}

func fileTail(file string, start bool) (*tail.Tail, error) {
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

	return tail.TailFile(file, config)
}
