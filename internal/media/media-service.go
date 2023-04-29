package media

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/egfanboy/mediapire-common/exceptions"
	"github.com/egfanboy/mediapire-media-host/internal/app"
	"github.com/egfanboy/mediapire-media-host/pkg/types"
	"github.com/google/uuid"

	"github.com/rs/zerolog/log"
)

type mediaService struct {
	app *app.App
}

var mediaTypeFactory = map[string]mediaFactory{
	"mp3": mp3Factory,
}

var mediaCache = map[string][]types.MediaItem{}

var mediaLookup = map[uuid.UUID]string{}

func unwrapCache() (unwrappedItems []types.MediaItem) {
	for _, items := range mediaCache {
		unwrappedItems = append(unwrappedItems, items...)
	}

	return
}

func (s *mediaService) GetMedia(ctx context.Context, mediaTypes []string) (items []types.MediaItem, err error) {
	items = make([]types.MediaItem, 0)

	if len(mediaCache) == 0 {
		err = errors.New("no media found")
		return
	}

	unwrappedCache := unwrapCache()

	if len(mediaTypes) == 0 {
		groupingFuncs := getGroupingFactories(s.app.FileTypes...)

		items = unwrappedCache

		for _, fn := range groupingFuncs {
			items = fn(items)
		}

		return

	}

	for _, item := range unwrappedCache {
		mediaTypesToCheck := strings.Join(mediaTypes, ",")
		if strings.Contains(mediaTypesToCheck, item.Extension) {
			items = append(items, item)
		}

	}

	groupingFuncs := getGroupingFactories(mediaTypes...)

	for _, fn := range groupingFuncs {
		items = fn(items)
	}

	return
}

func (s *mediaService) ScanDirectories(directories ...string) (err error) {
	failed := make([]string, 0)

	for _, d := range directories {
		err = s.ScanDirectory(d)
		if err != nil {
			log.Error().Err(err).Str("directory", d)

			failed = append(failed, d)
		}
	}

	if len(failed) > 0 {
		err = fmt.Errorf("failed to scan the following directories: %s", strings.Join(failed, ", "))
	}

	return
}

func (s *mediaService) ScanDirectory(directory string) (err error) {
	items := make([]types.MediaItem, 0)

	wg := sync.WaitGroup{}

	err = filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Warn().Err(err).Msgf("Error occured when walking %s, skipping.", path)
			return filepath.SkipDir
		}

		ext := filepath.Ext(path)

		// filepath.Ext returns . before file type, strip it away
		if strings.HasPrefix(ext, ".") {
			ext = strings.ReplaceAll(ext, ".", "")
		}

		if s.app.IsMediaSupported(ext) {
			if factory, ok := mediaTypeFactory[ext]; ok {
				wg.Add(1)

				go func() {
					defer wg.Done()

					item, err := factory(path, ext, info)

					if err != nil {
						log.Error().Err(err).Str("file", info.Name())
						return
					}

					item.Path = path
					item.Id = uuid.New()

					items = append(items, item)
				}()

			} else {
				log.Warn().Msgf("No factory for supported media type %s, cannot parse file.", ext)
			}
		}

		return nil
	})

	wg.Wait()

	// Add to cache
	mediaCache[directory] = items

	return
}

func (s *mediaService) StreamMedia(ctx context.Context, id uuid.UUID) ([]byte, error) {
	filePath, err := s.getFilePathFromId(ctx, id)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(filePath)
	if err != nil && errors.Is(err, os.ErrInvalid) || errors.Is(err, os.ErrNotExist) {
		return nil, &exceptions.ApiException{Err: err, StatusCode: http.StatusNotFound}
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	b := make([]byte, fileInfo.Size())

	_, err = file.Read(b)
	return b, err
}

func (s *mediaService) getFilePathFromId(ctx context.Context, id uuid.UUID) (string, error) {
	if path, ok := mediaLookup[id]; ok {
		return path, nil
	}

	for _, item := range unwrapCache() {
		if item.Id == id {
			mediaLookup[id] = item.Path
			return item.Path, nil
		}
	}

	return "", &exceptions.ApiException{Err: fmt.Errorf("no item with id %s", id.String()), StatusCode: http.StatusNotFound}
}

func (s *mediaService) UnsetDirectory(directory string) error {
	delete(mediaCache, directory)

	return nil
}

func NewMediaService() MediaApi {
	return &mediaService{app: app.GetApp()}
}
