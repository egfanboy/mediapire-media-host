package media

import (
	"github.com/dhowden/tag"
)

type Mp3Metadata struct {
	Artist     string
	Title      string
	Album      string
	Genre      string
	TrackIndex int
	TrackOf    int
	Length     float64
}

func mp3MetadataFromTag(m tag.Metadata) Mp3Metadata {
	trackIndex, trackOf := m.Track()
	return Mp3Metadata{
		Artist:     m.Artist(),
		Title:      m.Title(),
		Album:      m.Album(),
		Genre:      m.Genre(),
		TrackIndex: trackIndex,
		TrackOf:    trackOf,
	}

}
