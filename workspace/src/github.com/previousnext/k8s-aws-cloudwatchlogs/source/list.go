package source

import (
	"os"
	"path/filepath"

	"github.com/previousnext/k8s-aws-cloudwatchlogs/k8slog"
)

// List returns a list of existing log files.
func List(dir string) ([]os.FileInfo, error) {
	var files []os.FileInfo

	err := filepath.Walk(dir, func(path string, file os.FileInfo, err error) error {
		if !k8slog.Validate(file) {
			return nil
		}

		files = append(files, file)

		return nil
	})
	if err != nil {
		return files, err
	}

	return files, nil
}
