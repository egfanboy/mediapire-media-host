package media

import "context"

type MediaApi interface {
	GetMedia(ctx context.Context) ([]MediaItem, error)
	ScanDirectory(directory string) error
	ScanDirectories(directories ...string) error
}
