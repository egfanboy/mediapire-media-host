package media

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"

	"github.com/dhowden/tag"
	"github.com/egfanboy/mediapire-media-host/internal/app"
	"github.com/egfanboy/mediapire-media-host/pkg/types"
	"github.com/rs/zerolog/log"
)

var mp3ItemChan = make(chan types.MediaItem, 10)

const (
	artFileExtension = "jpg"
)

func getArtPathForMp3Item(item types.MediaItem) (string, error) {
	metadata := item.Metadata.(Mp3Metadata)

	return path.Join(app.GetApp().ArtPath, fmt.Sprintf("%s-%s.%s", metadata.Artist, metadata.Album, artFileExtension)), nil
}

func getArtForMp3Item(item types.MediaItem) ([]byte, error) {
	mp3File, err := os.OpenFile(item.Path, 0, fs.FileMode(os.O_RDONLY))
	if err != nil {
		return nil, err
	}

	defer mp3File.Close()

	rs := io.ReadSeeker(mp3File)

	m, err := tag.ReadFrom(rs)
	if err != nil {
		return nil, err
	}

	return m.Picture().Data, nil
}

func processMp3Item() {
	for value := range mp3ItemChan {
		log.Debug().Msgf("handling mp3 File: %s", value.Path)

		artPath, err := getArtPathForMp3Item(value)
		if err != nil {
			continue
		}

		_, err = os.ReadFile(artPath)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			log.Err(err).Msgf("could not read file for %s", value.Name)
			continue
		}

		if errors.Is(err, os.ErrNotExist) {
			log.Debug().Msgf("writing album art for to disk mp3 file: %s", value.Path)

			artBytes, err := getArtForMp3Item(value)
			if err != nil {
				log.Err(err).Msgf("could not get art for mp3 item %s", value.Name)
			}

			err = os.WriteFile(artPath, artBytes, os.ModePerm)
			if err != nil {
				continue
			}
		}

	}
}

func init() {
	go processMp3Item()
}
