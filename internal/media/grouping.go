package media

import (
	"sort"

	"github.com/egfanboy/mediapire-media-host/pkg/types"

	"github.com/rs/zerolog/log"
)

type mediaGroupingFactory func(items []types.MediaItem) []types.MediaItem

var groupingFactories = map[string][]mediaGroupingFactory{
	"mp3": {mp3AlbumGrouper},
}

func getGroupingFactories(mediaTypes ...string) []mediaGroupingFactory {

	funcs := make([]mediaGroupingFactory, 0)

	for _, mediaType := range mediaTypes {
		if typeFns, ok := groupingFactories[mediaType]; ok {

			funcs = append(funcs, typeFns...)
		} else {
			log.Debug().Msgf("No grouping function(s) found for media type %s", mediaType)
		}
	}

	return funcs
}

// Groups songs by albums and returns them in alphabetical order based on album
func mp3AlbumGrouper(items []types.MediaItem) []types.MediaItem {
	albumMap := make(map[string][]*types.MediaItem)

	for i := range items {
		item := items[i]
		if item.Extension != "mp3" {
			continue
		}

		metadata := item.Metadata.(Mp3Metadata)

		albumTitle := metadata.Album

		trackOf := metadata.TrackOf

		if album, ok := albumMap[albumTitle]; !ok {
			// we have metadata for number of tracks, create a slice of that size and put the songs in the proper order
			if trackOf != 0 {
				album := make([]*types.MediaItem, trackOf)

				album[metadata.TrackIndex-1] = &item

				albumMap[albumTitle] = album
			} else {
				// create new key for album with the current item as the first track
				albumMap[albumTitle] = []*types.MediaItem{&item}
			}
		} else {
			// we have metadata for number of tracks, add track in proper spot
			if trackOf != 0 {
				album[metadata.TrackIndex-1] = &item
			} else {
				album = append(album, &item)
			}

			albumMap[albumTitle] = album
		}

	}

	albums := make([]string, 0)

	for k := range albumMap {
		albums = append(albums, k)
	}

	sort.Strings(albums)

	itemsGroupedByAlbum := make([]types.MediaItem, 0)

	for _, album := range albums {
		albumItems := albumMap[album]
		for _, item := range albumItems {
			if item != nil {
				itemsGroupedByAlbum = append(itemsGroupedByAlbum, *item)
			}
		}
	}

	return itemsGroupedByAlbum
}
