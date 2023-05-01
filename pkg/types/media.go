package types

import "github.com/google/uuid"

type MediaItem struct {
	Name      string      `json:"name"`
	Extension string      `json:"extenstion"`
	Path      string      `json:"-"`
	Id        uuid.UUID   `json:"id"`
	Metadata  interface{} `json:"metadata"`
}

type DownloadRequest []uuid.UUID
