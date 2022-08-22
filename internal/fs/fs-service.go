package fs

import (
	"os"
	"path/filepath"

	"github.com/egfanboy/mediapire-media-host/internal/media"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
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
			log.Error().Err(err).Msgf("Failed to create new watcher for directory: %s", directory)

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
						log.Error().Err(err).Msgf("Failed to scan directory: %s", directory)
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Error().Err(err).Msgf("An error occured in the watcher for directory %s", directory)
			}
		}

	}()

	// Errors are handled internally
	filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {

		// If there is an error, just skip since we don't want one error to shutdown everything
		if err != nil {
			log.Warn().Err(err).Msgf("Error occured when walking %s, skipping.", path)

			return filepath.SkipDir
		}

		if info.IsDir() {
			err = watcher.Add(path)

			if err != nil {
				log.Error().Err(err).Msgf("Failed to start watching directory: %s", path)
			}
		}

		return nil
	})

	return nil
}

func (s *fsService) CloseWatchers() {
	for _, w := range watcherMapping {
		w.Close()
	}
}

func NewFsService() (FsApi, error) {

	return &fsService{mediaService: media.NewMediaService()}, nil
}
