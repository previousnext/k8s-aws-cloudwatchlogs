package source

import (
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/previousnext/k8s-aws-cloudwatchlogs/k8slog"
	"github.com/radovskyb/watcher"
)

// Watch streams new files added to the directory.
func Watch(dir string) (chan os.FileInfo, error) {
	files := make(chan os.FileInfo)

	// Now we want to start watching for new files which get added to a directory.
	w := watcher.New()

	// SetMaxEvents to 1 to allow at most 1 event's to be received
	// on the Event channel per watching cycle.
	//
	// If SetMaxEvents is not set, the default is to send all events.
	w.SetMaxEvents(1)

	// Only notify create events.
	w.FilterOps(watcher.Create)

	go func() {
		for {
			select {
			case event := <-w.Event:
				if !k8slog.Validate(event.FileInfo) {
					break
				}

				// Pass the file back so others can inspect the file.
				files <- event.FileInfo
			case err := <-w.Error:
				log.Fatal(err)
			case <-w.Closed:
				return
			}
		}
	}()

	// Watch this folder for changes.
	err := w.Add(dir)
	if err != nil {
		return files, err
	}

	// Start the watching process - it'll check for changes every 100ms.
	go func() {
		err = w.Start(time.Millisecond * 100)
		if err != nil {
			log.Fatal(err)
		}
	}()

	return files, err
}
