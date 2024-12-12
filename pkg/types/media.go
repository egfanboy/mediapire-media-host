package types

import "github.com/google/uuid"

type MediaItem struct {
	Name      string      `json:"name"`
	Extension string      `json:"extension"`
	Path      string      `json:"-"`
	Id        uuid.UUID   `json:"id"`
	Metadata  interface{} `json:"metadata"`
	// Top level directory this item belongs to
	ParentDir string `json:"-"`
}

type DownloadRequest []uuid.UUID
