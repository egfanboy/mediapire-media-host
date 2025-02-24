package types

type MediaItem struct {
	Name      string      `json:"name"`
	Extension string      `json:"extension"`
	Path      string      `json:"-"`
	Id        string      `json:"id"`
	Metadata  interface{} `json:"metadata"`
	// Direct parent of the item
	ParentDir string `json:"-"`
}

type DownloadRequest []string
