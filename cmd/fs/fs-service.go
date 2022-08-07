package fs

import (
	"fmt"
	"log"
	"mediapire-media-host/cmd/media"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type fsService struct {
	mediaService media.MediaApi
}

var watcherMapping = map[string]*fsnotify.Watcher{}

func (s *fsService) WatchDirectory(directory string) error {
	var watcher *fsnotify.Watcher

	if w, ok := watcherMapping[directory]; !ok {

		newW, err := fsnotify.NewWatcher()

		if err != nil {
			log.Fatal("NewWatcher failed: ", err)

			return err
		}

		watcher = newW
	} else {
		watcher = w
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				// Only ignore permission change events
				if event.Op != fsnotify.Chmod {
					//  TODO: only scan directory that changed
					err := s.mediaService.ScanDirectory(directory)
					if err != nil {
						fmt.Printf("Failed to scan directory: %s", directory)
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}

	}()

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {

		// If there is an error, just skip since we don't want one error to shutdown everything
		if err != nil {
			fmt.Printf("Failed to start walking directory: %s", path)
			return filepath.SkipDir
		}

		if info.IsDir() {
			err = watcher.Add(path)

			if err != nil {
				fmt.Printf("Failed to start watching directory: %s", path)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Print("Failed to start watchers for directory: " + directory)
	}

	return err
}

func (s *fsService) CloseWatchers() {
	for _, w := range watcherMapping {
		w.Close()
	}
}

func NewFsService() (FsApi, error) {

	return &fsService{mediaService: media.NewMediaService()}, nil
}
