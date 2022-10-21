package media

import (
	"context"

	"github.com/egfanboy/mediapire-media-host/pkg/types"
)

type MediaApi interface {
	GetMedia(ctx context.Context, mediaTypes []string) ([]types.MediaItem, error)
	ScanDirectory(directory string) error
	ScanDirectories(directories ...string) error
	StreamMedia(ctx context.Context, fileName string) ([]byte, error)
}
