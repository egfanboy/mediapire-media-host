package types

type MediaItem struct {
	Name      string      `json:"name"`
	Extension string      `json:"extenstion"`
	Path      string      `json:"path"`
	Metadata  interface{} `json:"metadata"`
}
