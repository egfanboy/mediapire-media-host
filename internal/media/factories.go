package media

import (
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/egfanboy/mediapire-media-host/pkg/types"

	"github.com/dhowden/tag"
	"github.com/tcolgate/mp3"
)

type mediaFactory func(path string, ext string, info os.FileInfo) (item types.MediaItem, err error)

func mp3Factory(path string, ext string, info os.FileInfo) (item types.MediaItem, err error) {

	f, err := os.OpenFile(path, 0, fs.FileMode(os.O_RDONLY))

	if err != nil {
		return
	}

	s := io.ReadSeeker(f)

	m, err := tag.ReadFrom(s)

	if err != nil {
		return
	}

	item.Name = strings.Replace(info.Name(), "."+ext, "", 1)
	item.Extension = ext
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

	return
}
