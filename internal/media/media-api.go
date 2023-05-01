package media

import (
	"context"

	"github.com/egfanboy/mediapire-media-host/pkg/types"
	"github.com/google/uuid"
)

type MediaApi interface {
	GetMedia(ctx context.Context, mediaTypes []string) ([]types.MediaItem, error)
	ScanDirectory(directory string) error
	ScanDirectories(directories ...string) error
	StreamMedia(ctx context.Context, id uuid.UUID) ([]byte, error)
	UnsetDirectory(directory string) error
	DownloadMedia(ctx context.Context, ids []uuid.UUID) ([]byte, error)
}
