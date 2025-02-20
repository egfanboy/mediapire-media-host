package media

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/egfanboy/mediapire-media-host/pkg/types"
	"github.com/google/uuid"

	"github.com/dhowden/tag"
	"github.com/tcolgate/mp3"
)

type mediaFactory func(path string, ext string) (item types.MediaItem, err error)

var mediaTypeFactory = map[string]mediaFactory{
	"mp3": mp3Factory,
}

func getFactory(ext string) mediaFactory {
	if factory, ok := mediaTypeFactory[ext]; !ok {
		return baseFactory
	} else {
		return factory
	}
}

func baseFactory(path, ext string) (item types.MediaItem, err error) {
	item.Path = path
	item.Extension = ext
	item.ParentDir = filepath.Dir(path)
	item.Name = strings.Replace(filepath.Base(path), "."+ext, "", 1)
	item.Id = uuid.New().String()

	return
}

func mp3Factory(path string, ext string) (item types.MediaItem, err error) {
	item, err = baseFactory(path, ext)
	if err != nil {
		return
	}

	f, err := os.OpenFile(path, 0, fs.FileMode(os.O_RDONLY))
	if err != nil {
		return
	}

	defer f.Close()

	s := io.ReadSeeker(f)

	m, err := tag.ReadFrom(s)
	if err != nil {
		return
	}

	metadata := mp3MetadataFromTag(m)

	d := mp3.NewDecoder(f)
	songTime := 0.0
	var frame mp3.Frame
	skipped := 0

	for {
		errD := d.Decode(&frame, &skipped)
		if errD != nil {
			if errD == io.EOF {
				break
			}

			return
		}

		songTime = songTime + frame.Duration().Seconds()
	}

	metadata.Length = songTime
	item.Metadata = metadata

	item.Id = hashMp3FileForId(item)

	return
}

// encodes metadata into a hash string and return last 12 digits to use as an id
func hashMp3FileForId(item types.MediaItem) string {
	metadata := item.Metadata.(Mp3Metadata)

	data := fmt.Sprintf("%s-%s-%s-%s-%d", item.ParentDir, metadata.Artist, metadata.Album, metadata.Title, metadata.TrackOf)
	hash := sha256.Sum256([]byte(data))
	hashStr := hex.EncodeToString(hash[:])

	return hashStr[len(hashStr)-12:]
}
