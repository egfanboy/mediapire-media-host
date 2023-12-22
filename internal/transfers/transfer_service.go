package transfers

import (
	"context"
	"os"
	"path"

	"github.com/egfanboy/mediapire-media-host/internal/app"
	"github.com/rs/zerolog/log"
)

type transfersApi interface {
	Download(ctx context.Context, transferId string) ([]byte, error)
}

type transfersService struct {
}

func (s *transfersService) Download(ctx context.Context, transferId string) ([]byte, error) {
	fileContent, err := os.ReadFile(path.Join(app.GetApp().DownloadPath, transferId+".zip"))
	if err != nil {
		log.Err(err).Msgf("Failed to open item for transfer with id %s", transferId)
		return nil, err
	}

	return fileContent, nil
}

func newTransfersService() transfersApi {
	return &transfersService{}
}
