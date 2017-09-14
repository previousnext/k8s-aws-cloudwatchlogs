package source

import (
	"os"

	"github.com/gobwas/glob"
)

const format = "*_*_*-*.log"

// Valid is used to check that our file has the correct naming strategy.
func Valid(file os.FileInfo) bool {
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
