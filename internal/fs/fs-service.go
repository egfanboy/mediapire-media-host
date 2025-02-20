package fs

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bep/debounce"
	"github.com/egfanboy/mediapire-media-host/internal/media"
	"github.com/egfanboy/mediapire-media-host/internal/utils"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
)

type fsService struct {
	mediaService media.MediaApi
}

var ignoredFiles = []string{".DS_Store"}

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

				err := s.handleWatcherEvent(event, directory, watcher)

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
			debouncer := debounce.New(time.Millisecond * 500)

			debouncer(func() {
				// Note: cannot use stat here since the file no longer exists on the system

				// remove the top level directory from the event name since it cntains the full path
				relativeDeletedFilePath := strings.Replace(event.Name, topLevelDirectory, "", 1)

				// trim / if it is there to have the format dir?/file
				relativeDeletedFilePath = strings.TrimPrefix(relativeDeletedFilePath, "/")

				parentDir := filepath.Dir(relativeDeletedFilePath)

				if parentDir != "." {
					err := watcher.Remove(event.Name)
					if err != nil {
						log.Error().Err(err).Msgf("failed to remove %s from the watchlist", event.Name)
					}

					s.mediaService.UnsetDirectory(event.Name)
					return
				}

				return
			})

		}

	case fsnotify.Chmod:
		// ignore
		return nil

	default:
		{
			if utils.Contains(ignoredFiles, filepath.Base(event.Name)) {
				log.Debug().Msg("Ignoring fs event since the target is an ignored file.")
				return nil
			}

			debouncer := debounce.New(time.Millisecond * 500)
			debouncer(func() {
				stat, err := os.Stat(event.Name)
				if err != nil {
					return
				}

				if event.Op == fsnotify.Create && stat.IsDir() {
					err := watcher.Add(event.Name)
					if err != nil {
						log.Err(err).Msgf("failed to add %s to the watchlist", event.Name)
					}
				}

				var dirToScan string

				if stat.IsDir() {
					dirToScan = event.Name
				} else {
					dirToScan = filepath.Dir(event.Name)
				}

				log.Debug().Msgf("Detected a change in directory %s, scanning it again for new media", dirToScan)
				s.mediaService.ScanDirectory(dirToScan)
			})

		}
	}
	return nil
}

func NewFsService() (FsApi, error) {

	return &fsService{mediaService: media.NewMediaService()}, nil
}
