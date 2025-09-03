package fs

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/egfanboy/mediapire-common/messaging"
	"github.com/egfanboy/mediapire-media-host/internal/app"
	"github.com/egfanboy/mediapire-media-host/internal/fs/ignorelist"
	"github.com/egfanboy/mediapire-media-host/internal/media"
	"github.com/egfanboy/mediapire-media-host/internal/rabbitmq"
	"github.com/egfanboy/mediapire-media-host/internal/utils"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
)

type fsService struct {
	mediaService media.MediaApi
}

const (
	// How often we process fs event batches
	eventProcessingInterval = time.Second * 10
	// Maximum amount of events
	eventBuffer = 100
)

var (
	ignoredFiles = []string{".DS_Store"}
)

type fsWatcher struct {
	w                            *fsnotify.Watcher
	deleteBatchProcessor         utils.AsyncBatchProcessor[fsnotify.Event]
	nonDestructiveBatchProcessor utils.AsyncBatchProcessor[fsnotify.Event]
	directory                    string
}

func (w *fsWatcher) ProcessEvents(events []fsnotify.Event) {
	log.Debug().Msgf("Processing %d events", len(events))

	affectedDirectories := utils.NewUnorderedSet[string]()

	for _, event := range events {
		stat, err := os.Stat(event.Name)
		if err != nil {
			return
		}

		if event.Op == fsnotify.Create && stat.IsDir() {
			err := w.w.Add(event.Name)
			if err != nil {
				log.Err(err).Msgf("failed to add %s to the watchlist", event.Name)
			}
		}

		if stat.IsDir() {
			affectedDirectories.Add(event.Name)
		} else {
			affectedDirectories.Add(filepath.Dir(event.Name))
		}

		mediaService := media.NewMediaService()
		for _, directory := range affectedDirectories.Values() {
			relativePath := strings.ReplaceAll(directory, w.directory, "")

			log.Debug().Msgf("Detected a change in directory %s, scanning it again for new media", relativePath)
			err := mediaService.ScanDirectory(directory)
			if err != nil {
				log.Err(err).Msgf("Failed to scan content in directory %s", relativePath)
			} else {
				w.sendMediaUpdateMessage()
			}

		}
	}
}

func (w *fsWatcher) ProcessDeletedItems(events []fsnotify.Event) {
	log.Debug().Msgf("Processing %d delete events", len(events))

	eventNames := utils.NewUnorderedSet[string]()

	// Get unique name of events
	for _, event := range events {
		eventNames.Add(event.Name)
	}

	affectedDirectories := utils.NewUnorderedSet[string]()

	// Get unique name of affected directories (parent of events)
	for _, event := range eventNames.Values() {
		affectedDirectories.Add(filepath.Dir(event))
	}

	// Try to remove items from media service
	media.NewMediaService().HandleFileSystemDeletions(context.Background(), eventNames.Values())

	for _, affectedDir := range affectedDirectories.Values() {
		relativePath := strings.ReplaceAll(affectedDir, w.directory, "")

		if utils.Contains(w.w.WatchList(), affectedDir) {
			content, err := os.ReadDir(affectedDir)
			if err != nil {
				if os.IsNotExist(err) {
					log.Debug().Msgf("Directory %s is no longer on disk, stop watching it.", relativePath)
					err := w.w.Remove(affectedDir)
					if err != nil {
						log.Err(err).Msgf("Failed to remove directory from watchlist %s", relativePath)
					}
				}
				continue
			}

			// Filter out ignored files and if no more content in this directory, stop watching and delete it from disk
			if len(utils.Filter(content, func(file fs.DirEntry) bool {
				return !utils.Contains(ignoredFiles, file.Name())
			})) == 0 {
				log.Debug().Msgf("Directory %s has no more content, stop watching it and delete it.", relativePath)
				err := w.w.Remove(affectedDir)
				if err != nil {
					log.Err(err).Msgf("Failed to remove directory from watchlist %s", relativePath)
				}

				err = os.RemoveAll(affectedDir)
				if err != nil {
					log.Err(err).Msgf("Failed to remove directory from disk %s", relativePath)
				}
			}
		}
	}

	w.sendMediaUpdateMessage()
}

func (w *fsWatcher) sendMediaUpdateMessage() {
	err := rabbitmq.PublishMessage(context.Background(), messaging.TopicNodeMediaChanged, messaging.NodeReadyMessage{Name: app.GetApp().Name, Id: app.GetApp().NodeId})
	if err != nil {
		log.Err(err).Msg("Failed to send media update message")
	}
}

func (w *fsWatcher) Stop() {
	w.deleteBatchProcessor.Stop()
	w.w.Close()
}

var watcherMapping = map[string]*fsWatcher{}

/*
Creates a fsnotify.Watcher for a given directory and walks the directory
adding subdirectories to the watch list and starting a goroutine to consume the fs events
*/
func (s *fsService) WatchDirectory(directory string) error {
	var watcher *fsWatcher

	if w, ok := watcherMapping[directory]; !ok {
		newW, err := fsnotify.NewWatcher()
		if err != nil {
			log.Error().Err(err).Msgf("Failed to create new watcher for directory: %s", directory)

			return err
		}

		watcher = &fsWatcher{
			w:         newW,
			directory: directory,
		}
		watcher.deleteBatchProcessor = utils.NewAsyncBatchProcessor(eventProcessingInterval, eventBuffer, watcher.ProcessDeletedItems)
		watcher.nonDestructiveBatchProcessor = utils.NewAsyncBatchProcessor(eventProcessingInterval, eventBuffer, watcher.ProcessEvents)
		watcher.w.Add(directory)
	} else {
		watcher = w
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.w.Events:
				if !ok {
					return
				}

				err := s.handleWatcherEvent(event, watcher)

				if err != nil {
					log.Error().Err(err).Msgf("Failed to handle change for directory: %s", directory)
				}
			case err, ok := <-watcher.w.Errors:
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
			err = watcher.w.Add(path)
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
		w.Stop()
	}
}

// topLevelDirectory refers to the directory item in the config
func (s *fsService) handleWatcherEvent(event fsnotify.Event, watcher *fsWatcher) error {
	if utils.Contains(ignoredFiles, filepath.Base(event.Name)) {
		log.Debug().Msg("Ignoring fs event since the target is an ignored file.")
		return nil
	}

	if ignorelist.GetIgnoreList().IsFileIgnored(event.Name) {
		log.Debug().Msg("Ignoring fs event since the target is a temporarily ignored file.")

		return nil
	}

	switch event.Op {
	case fsnotify.Rename | fsnotify.Remove:
		watcher.deleteBatchProcessor.Add(event)
	case fsnotify.Remove:
		watcher.deleteBatchProcessor.Add(event)
	case fsnotify.Rename:
		{
			/**
			* On MacOS if a user moves a file to trash (deletes it) we get a Rename Op
			* Therefore, try to stat the file and if the error is the not exist error add it to the delete batch
			**/
			_, err := os.Stat(event.Name)
			if err != nil {
				if os.IsNotExist(err) {
					watcher.deleteBatchProcessor.Add(event)
					return nil
				}

				return err
			}
		}

	case fsnotify.Chmod:
		// ignore
		return nil

	default:
		watcher.nonDestructiveBatchProcessor.Add(event)

	}
	return nil
}

func NewFsService() (FsApi, error) {
	return &fsService{mediaService: media.NewMediaService()}, nil
}
