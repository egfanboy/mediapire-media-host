package media

import (
	"github.com/dhowden/tag"
)

type Mp3Metadata struct {
	Artist     string  `json:"artist"`
	Title      string  `json:"title"`
	Album      string  `json:"album"`
	Genre      string  `json:"genre"`
	TrackIndex int     `json:"trackIndex"`
	TrackOf    int     `json:"trackOf"`
	Length     float64 `json:"length"`
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
