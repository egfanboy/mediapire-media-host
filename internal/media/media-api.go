package media

import (
	"context"

	"github.com/egfanboy/mediapire-media-host/pkg/types"
)

type MediaApi interface {
	GetMedia(ctx context.Context, mediaTypes []string) ([]types.MediaItem, error)
	ScanDirectory(directory string) error
	ScanDirectories(directories ...string) error
	StreamMedia(ctx context.Context, id string) ([]byte, error)
	UnsetDirectory(directory string) error
	DownloadMedia(ctx context.Context, ids []string) ([]byte, error)
	DeleteMedia(ctx context.Context, ids []string) error
	CleanupDownloadContent(ctx context.Context, transferId string) error
	GetMediaArt(ctx context.Context, id string) ([]byte, error)
	HandleFileSystemDeletions(ctx context.Context, files []string) error
	UpdateItem(ctx context.Context, id string, newContent []byte) (types.MediaItem, error)
}
