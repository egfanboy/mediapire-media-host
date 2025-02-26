package media

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/egfanboy/mediapire-common/exceptions"
	"github.com/egfanboy/mediapire-media-host/internal/app"
	"github.com/egfanboy/mediapire-media-host/internal/fs/ignorelist"
	"github.com/egfanboy/mediapire-media-host/internal/utils"
	"github.com/egfanboy/mediapire-media-host/pkg/types"

	"github.com/rs/zerolog/log"
)

type mediaService struct {
	app *app.App
}

var mediaCache = utils.NewConcurrentMap[string, []types.MediaItem]()

var mediaLookup = utils.NewConcurrentMap[string, types.MediaItem]()

func unwrapCache() (unwrappedItems []types.MediaItem) {
	for _, items := range mediaCache.Get() {
		unwrappedItems = append(unwrappedItems, items...)
	}

	return
}

func (s *mediaService) GetMedia(ctx context.Context, mediaTypes []string) (items []types.MediaItem, err error) {
	items = make([]types.MediaItem, 0)

	if mediaCache.Len() == 0 {
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

func (s *mediaService) scanDirectory(directory string, wg *sync.WaitGroup, wp utils.WorkerPool, items chan<- types.MediaItem) error {
	defer wg.Done()

	visit := func(path string, info os.FileInfo, err error) error {
		if err != nil && err != os.ErrNotExist {
			return err
		}

		// ignore itself to avoid infinite loop
		if info.Mode().IsDir() && path != directory {
			wg.Add(1)
			go s.scanDirectory(path, wg, wp, items)
			// this will skip this directory since we spawn a goroutine to handle it
			return filepath.SkipDir
		}

		if info.Mode().IsRegular() && info.Size() > 0 {
			wg.Add(1)
			go s.processFile(path, wg, wp, items)
		}

		return nil
	}

	wp.Work()

	defer wp.Done()

	return filepath.Walk(directory, visit)
}

func (s *mediaService) processFile(path string, wg *sync.WaitGroup, wp utils.WorkerPool, items chan<- types.MediaItem) {
	defer wg.Done()

	ext := filepath.Ext(path)

	// filepath.Ext returns . before file type, strip it away
	if strings.HasPrefix(ext, ".") {
		ext = strings.ReplaceAll(ext, ".", "")
	}

	if !s.app.IsMediaSupported(ext) {
		return
	}

	wp.Work()
	defer wp.Done()

	factory := getFactory(ext)

	item, err := factory(path, ext)
	if err != nil {
		log.Err(err).Msgf("failed to create media item for file %s", filepath.Base(path))
		return
	}

	items <- item
}

func (s *mediaService) processItems(items <-chan types.MediaItem, result chan<- map[string][]types.MediaItem) {
	mediaItems := map[string][]types.MediaItem{}

	for item := range items {
		// TODO: Make this dynamic for extension type
		if item.Extension == "mp3" {
			mp3ItemChan <- item
		}

		mediaItems[item.ParentDir] = append(mediaItems[item.ParentDir], item)
	}

	result <- mediaItems
}

func (s *mediaService) ScanDirectory(directory string) (err error) {
	workers := 2 * runtime.GOMAXPROCS(0)

	wp := utils.NewWorkerPool(workers)
	items := make(chan types.MediaItem)
	result := make(chan map[string][]types.MediaItem)

	wg := new(sync.WaitGroup)

	go s.processItems(items, result)

	wg.Add(1)

	// will walk through directories and spawn goroutines to handle subdirectories and files
	s.scanDirectory(directory, wg, wp, items)

	wg.Wait()

	// Close items channel since all files have been processed at this point
	close(items)

	results := <-result
	for k, v := range results {
		mediaCache.Add(k, v)
	}

	return
}

func (s *mediaService) StreamMedia(ctx context.Context, id string) ([]byte, error) {
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

func (s *mediaService) getFilePathFromId(ctx context.Context, id string) (string, error) {
	item, err := s.getMediaItemFromId(ctx, id)
	if err != nil {
		return "", err
	}

	return item.Path, nil
}

func (s *mediaService) getMediaItemFromId(ctx context.Context, id string) (types.MediaItem, error) {
	if item, ok := mediaLookup.GetKey(id); ok {
		return item, nil
	}

	for _, item := range unwrapCache() {
		if item.Id == id {
			mediaLookup.Add(id, item)
			return item, nil
		}
	}

	return types.MediaItem{}, &exceptions.ApiException{Err: fmt.Errorf("no item with id %s", id), StatusCode: http.StatusNotFound}
}

func (s *mediaService) UnsetDirectory(directory string) error {
	mediaCache.Delete(directory)

	return nil
}

func (s *mediaService) DownloadMedia(ctx context.Context, ids []string) ([]byte, error) {
	log.Info().Msg("Start: downloading media")
	buf := new(bytes.Buffer)

	zipWriter := zip.NewWriter(buf)

	items := make([]types.MediaItem, len(ids))

	for i, itemId := range ids {
		item, err := s.getMediaItemFromId(ctx, itemId)
		if err != nil {
			log.Err(err).Msgf("Failed to get item with id %q", itemId)
			return nil, err
		}

		items[i] = item
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

func (s *mediaService) DeleteMedia(ctx context.Context, ids []string) error {
	failedToDelete := make([]string, 0)

	for _, itemId := range ids {
		item, err := s.getMediaItemFromId(ctx, itemId)
		if err != nil {
			failedToDelete = append(failedToDelete, fmt.Sprintf("Failed to get item with id %q", itemId))

			continue
		}

		err = os.Remove(item.Path)
		if err != nil {
			failedToDelete = append(failedToDelete, fmt.Sprintf("Failed to delete item with id %q: %s", itemId, err.Error()))
		}

		err = s.removeItemFromCache(item)
		if err != nil {
			log.Err(err).Msg("Failed to remove item from the cache")
		}

	}

	if len(failedToDelete) > 0 {
		return fmt.Errorf("encountered the following errors during delete: %s", strings.Join(failedToDelete, "\n"))
	}

	return nil
}

func (s *mediaService) removeItemFromCache(item types.MediaItem) error {
	log.Info().Msgf("Removing item with id %q from the media cache", item.Id)

	// remove the item from the lookup
	mediaLookup.Delete(item.Id)

	if parentDirCache, ok := mediaCache.GetKey(item.ParentDir); !ok {
		return fmt.Errorf("parent dir for item %q is not in the cache", item.Id)
	} else {
		newCache := make([]types.MediaItem, 0)

		for _, cachedItem := range parentDirCache {
			// different item, add it to the new cache
			if cachedItem.Id != item.Id {
				newCache = append(newCache, cachedItem)
			}
		}

		mediaCache.Add(item.ParentDir, newCache)
	}

	return nil
}

func (s *mediaService) CleanupDownloadContent(ctx context.Context, transferId string) error {
	log.Info().Msgf("Deleting content for transfer with id %s", transferId)

	err := os.RemoveAll(path.Join(s.app.DownloadPath, transferId+".zip"))
	if err != nil {
		log.Err(err).Msgf("Failed to delete content for transfer with id %s", transferId)
	}

	return nil
}

func (s *mediaService) GetMediaArt(ctx context.Context, id string) ([]byte, error) {
	item, err := s.getMediaItemFromId(ctx, id)
	if err != nil {
		return nil, err
	}

	if item.Extension != "mp3" {
		return nil, &exceptions.ApiException{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("media art is not supported for media of type %s", item.Extension),
		}
	}

	// Assume only mp3 works from this point on, refactor for other types in the future
	artPath, err := getArtPathForMp3Item(item)
	if err != nil {
		return nil, err
	}

	artFile, err := os.OpenFile(artPath, 0, fs.FileMode(os.O_RDONLY))
	if err != nil {
		if os.IsNotExist(err) {
			log.Debug().Msgf("art for item %q not saved on disk, reading file to get album art", item.Id)

			art, err := getArtForMp3Item(item)
			if err != nil {
				errMsg := fmt.Sprintf("failed to open file for media item %q", item.Id)
				log.Err(err).Msg(errMsg)
				return nil, errors.New(errMsg)
			}
			// Send to channel so it can be processed and cached for the next run
			mp3ItemChan <- item

			return art, nil
		}

		// all other errors
		return nil, err
	}

	defer artFile.Close()

	buf := new(bytes.Buffer)

	_, err = io.Copy(buf, artFile)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *mediaService) HandleFileSystemDeletions(ctx context.Context, files []string) error {
	// file could be a specific file or a directory
	for _, file := range files {
		// handle scenario for update metadata where we overwrite the file which triggers a delete and create event
		// in this scenario the file was deleted at one point so we handle it here but we need to ensure if it still
		// exists we do not remove it if it not a directory
		stat, err := os.Stat(file)
		if err != nil {
			if !os.IsNotExist(err) {
				continue
			}
		}

		// need to check if err is not nil since still a valid code path from above if we get a specific error
		if err == nil && !stat.IsDir() {
			log.Debug().Msgf("file %s still exists, not removing it", file)
			continue
		}

		dir := filepath.Dir(file)

		if mediaForDirectory, ok := mediaCache.GetKey(dir); !ok {
			log.Debug().Msgf("Parent dir for file %s does not have any media", file)
			continue
		} else {

		mediaLoop:
			for _, media := range mediaForDirectory {
				if media.Path == file || media.ParentDir == file {
					err := s.removeItemFromCache(media)
					if err != nil {
						log.Err(err).Msgf("failed to remove delete item %s from cache", media.Name)
					}
					// found item for this file break and go to next file
					break mediaLoop
				}
			}
		}

	}

	return nil

}

func (s *mediaService) UpdateItem(ctx context.Context, id string, newContent []byte) (types.MediaItem, error) {
	item, err := s.getMediaItemFromId(ctx, id)
	if err != nil {
		return types.MediaItem{}, err
	}

	ignoreList := ignorelist.GetIgnoreList()
	ignoreList.AddFile(item.Path)
	defer ignoreList.RemoveFile(item.Path)

	err = os.WriteFile(item.Path, newContent, os.ModeType)
	if err != nil {
		log.Err(err).Msgf("Failed to write new content for file %s", item.Id)

		return types.MediaItem{}, err
	}

	// File is now changed, re-process it
	wp := utils.NewWorkerPool(1)
	items := make(chan types.MediaItem, 1)
	wg := new(sync.WaitGroup)

	wg.Add(1)
	s.processFile(item.Path, wg, wp, items)
	wg.Wait()

	newItem := <-items
	if item.Id != newItem.Id {
		if parentDirCache, ok := mediaCache.GetKey(item.ParentDir); !ok {
			return types.MediaItem{}, fmt.Errorf("parent dir for item %q is not in the cache", item.Id)
		} else {
			newCache := make([]types.MediaItem, len(parentDirCache))

			for i, cachedItem := range parentDirCache {
				// same item as the input, need to set it to the updated item
				if cachedItem.Id == item.Id {
					newCache[i] = newItem
				} else {
					newCache[i] = cachedItem
				}
			}

			mediaCache.Add(item.ParentDir, newCache)

			return newItem, nil
		}
	}

	return newItem, nil
}

func NewMediaService() MediaApi {
	return &mediaService{app: app.GetApp()}
}
