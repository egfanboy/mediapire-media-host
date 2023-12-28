package fs

import (
	"os"
	"path/filepath"
	"time"

	"github.com/bep/debounce"
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
		debouncer := debounce.New(time.Millisecond * 500)

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				var err error
				debouncer(func() {
					err = s.handleWatcherEvent(event, directory, watcher)
				})
				if err != nil {
					log.Error().Err(err).Msgf("Failed to handle change for directory: %s", directory)
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
		//  check if we just support 1 level of nested
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

// topLevelDirectory refers to the directory item in the config
func (s *fsService) handleWatcherEvent(event fsnotify.Event, topLevelDirectory string, watcher *fsnotify.Watcher) error {
	switch event.Op {
	case fsnotify.Remove:
		{
			err := watcher.Remove(event.Name)
			if err != nil {
				log.Error().Err(err).Msgf("failed to remove %s from the watchlist", event.Name)
			}

			return s.mediaService.UnsetDirectory(event.Name)
		}

	case fsnotify.Chmod:
		// ignore
		return nil

	default:
		stat, err := os.Stat(event.Name)
		if err != nil {
			return err
		}

		if event.Op == fsnotify.Create && stat.IsDir() {
			err := watcher.Add(event.Name)
			if err != nil {
				log.Error().Err(err).Msgf("failed to add %s to the watchlist", event.Name)
			}
		}

		var dirToScan string

		if stat.IsDir() {
			dirToScan = event.Name
		} else {
			dirToScan = filepath.Dir(event.Name)
		}

		log.Debug().Msgf("Detected a change in directory %s, scanning it again for new media", dirToScan)
		return s.mediaService.ScanDirectory(dirToScan)
	}
}

func NewFsService() (FsApi, error) {

	return &fsService{mediaService: media.NewMediaService()}, nil
}
