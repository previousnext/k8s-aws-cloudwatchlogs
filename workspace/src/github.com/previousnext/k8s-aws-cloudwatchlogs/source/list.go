package source

import (
	"os"
	"path/filepath"
)

// List returns a list of existing log files.
func List(dir string) ([]os.FileInfo, error) {
	var files []os.FileInfo

	err := filepath.Walk(dir, func(path string, file os.FileInfo, err error) error {
		if !Valid(file) {
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
