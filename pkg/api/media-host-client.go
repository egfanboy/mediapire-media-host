package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/egfanboy/mediapire-media-host/pkg/types"
)

type MediaHostApi interface {
	GetHealth(h types.Host) (*http.Response, error)
}

type mediaHostIntegration struct{}

func (i *mediaHostIntegration) GetHealth(h types.Host) (*http.Response, error) {
	hostUri := fmt.Sprintf("%s:%v", h.Host(), h.Port())

	return http.Get(fmt.Sprintf("%s://%s/api/v1/health", h.Scheme(), hostUri))
}

func NewIntegration(ctx context.Context) MediaHostApi {
	return &mediaHostIntegration{}
}
