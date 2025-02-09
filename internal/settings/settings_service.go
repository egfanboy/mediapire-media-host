package settings

import (
	"context"

	"github.com/egfanboy/mediapire-media-host/internal/app"
	"github.com/egfanboy/mediapire-media-host/pkg/types"
)

type settingsApi interface {
	GetSettings(ctx context.Context) (types.MediaHostSettings, error)
}

type settingsService struct {
}

func (s *settingsService) GetSettings(ctx context.Context) (types.MediaHostSettings, error) {
	return types.MediaHostSettings{FileTypes: app.GetApp().FileTypes}, nil
}

func newSettingsService() settingsApi {
	return &settingsService{}
}
