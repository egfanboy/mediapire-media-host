package media

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
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

var mediaLookup = map[uuid.UUID]types.MediaItem{}

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
	item, err := s.getMediaItemFromId(ctx, id)
	if err != nil {
		return "", err
	}

	return item.Path, nil
}

func (s *mediaService) getMediaItemFromId(ctx context.Context, id uuid.UUID) (types.MediaItem, error) {
	if item, ok := mediaLookup[id]; ok {
		return item, nil
	}

	for _, item := range unwrapCache() {
		if item.Id == id {
			mediaLookup[id] = item
			return item, nil
		}
	}

	return types.MediaItem{}, &exceptions.ApiException{Err: fmt.Errorf("no item with id %s", id.String()), StatusCode: http.StatusNotFound}
}

func (s *mediaService) UnsetDirectory(directory string) error {
	delete(mediaCache, directory)

	return nil
}

func (s *mediaService) DownloadMedia(ctx context.Context, ids []uuid.UUID) ([]byte, error) {
	log.Info().Msg("Start: downloading media")
	buf := new(bytes.Buffer)

	zipWriter := zip.NewWriter(buf)

	items := make([]types.MediaItem, len(ids))

	for _, itemId := range ids {
		item, err := s.getMediaItemFromId(ctx, itemId)
		if err != nil {
			log.Err(err).Msgf("Failed to get item with id %q", itemId)
			return nil, err
		}

		items = append(items, item)
	}

	groupingFuncs := getGroupingFactories(s.app.FileTypes...)

	for _, fn := range groupingFuncs {
		items = fn(items)
	}

	for _, item := range items {
		log.Debug().Msgf("adding item with id %q to archive", item.Id)

		file, err := os.Open(item.Path)
		if err != nil {
			log.Err(err).Msgf("Failed to open item with id %q", item.Id)
			return nil, err
		}

		itemPath := fmt.Sprintf("%s.%s", item.Name, item.Extension)

		// if the item we are handling is an MP3 file, save it as Album/song.mp3
		if metatada, ok := item.Metadata.(Mp3Metadata); ok {
			log.Debug().Msgf("item with id %q is an mp3 file, save it under an folder for the album", item.Id)
			itemPath = fmt.Sprintf("%s/%s.%s", metatada.Album, item.Name, item.Extension)
		}

		writer, err := zipWriter.Create(itemPath)
		if err != nil {
			log.Err(err)
			return nil, err
		}

		if _, err := io.Copy(writer, file); err != nil {
			log.Err(err).Msgf("Failed to copy file to archive for item with id %q", item.Id)
			return nil, err
		}

		file.Close()
	}

	err := zipWriter.Close()

	log.Info().Msg("Finished: downloading media")

	return buf.Bytes(), err
}

func NewMediaService() MediaApi {
	return &mediaService{app: app.GetApp()}
}
