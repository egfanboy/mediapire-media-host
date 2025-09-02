package media

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/egfanboy/mediapire-media-host/internal/utils"
	"github.com/egfanboy/mediapire-media-host/pkg/types"
	"github.com/google/uuid"

	"github.com/dhowden/tag"
	"github.com/tcolgate/mp3"
)

type mediaFactory func(path string, ext string, cache *utils.ConcurrentMap[string, string]) (item types.MediaItem, err error)

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

func baseFactory(path, ext string, cache *utils.ConcurrentMap[string, string]) (item types.MediaItem, err error) {
	item.Path = path
	item.Extension = ext
	item.ParentDir = filepath.Dir(path)
	item.Name = strings.Replace(filepath.Base(path), "."+ext, "", 1)

	if id, ok := cache.GetKey(item.Path); !ok {
		item.Id = uuid.New().String()
	} else {
		item.Id = id
	}

	return
}

func mp3Factory(path string, ext string, cache *utils.ConcurrentMap[string, string]) (item types.MediaItem, err error) {
	item, err = baseFactory(path, ext, cache)
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

			err = errD
			return
		}

		songTime = songTime + frame.Duration().Seconds()
	}

	metadata.Length = songTime
	item.Metadata = metadata
	/* Use the song title as the name of the item as opposed to the file name.
	** This is in case the name of the song was changed through metadata but the file name
	** remains the same.
	 */
	item.Name = metadata.Title

	return
}
