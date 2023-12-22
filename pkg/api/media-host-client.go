package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/egfanboy/mediapire-media-host/pkg/types"
	"github.com/google/uuid"
)

const (
	baseMediaPath     = "/api/v1/media"
	baseTransfersPath = "/api/v1/transfers"
)

type MediaHostApi interface {
	GetHealth(h types.Host) (*http.Response, error)
	GetMedia(h types.Host) ([]types.MediaItem, *http.Response, error)
	StreamMedia(h types.Host, mediaId uuid.UUID) ([]byte, *http.Response, error)
	DownloadTransfer(h types.Host, transferId string) ([]byte, *http.Response, error)
}

func buildUriFromHost(h types.Host, apiUri string) string {
	return fmt.Sprintf("%s://%s:%v%s", h.Scheme(), h.Host(), h.Port(), apiUri)

}

type mediaHostClient struct{}

func (c *mediaHostClient) GetHealth(h types.Host) (*http.Response, error) {

	return http.Get(buildUriFromHost(h, "/api/v1/health"))
}

func (c *mediaHostClient) GetMedia(h types.Host) (result []types.MediaItem, r *http.Response, err error) {

	r, err = http.Get(buildUriFromHost(h, baseMediaPath))
	if err != nil {
		return
	}

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &result)

	return
}

func (c *mediaHostClient) StreamMedia(h types.Host, mediaId uuid.UUID) (b []byte, r *http.Response, err error) {
	r, err = http.Get(buildUriFromHost(h, fmt.Sprintf("%s/stream?id=%q", baseMediaPath, mediaId)))
	if err != nil {
		return
	}

	b, err = io.ReadAll(r.Body)

	defer r.Body.Close()

	return
}

func (c *mediaHostClient) DownloadTransfer(h types.Host, transferId string) ([]byte, *http.Response, error) {
	r, err := http.Get(buildUriFromHost(h, fmt.Sprintf("%s/%s/download", baseTransfersPath, transferId)))
	if err != nil {
		return nil, r, err
	}

	if r.StatusCode >= 300 {
		return nil, r, fmt.Errorf("request failed with status %d", r.StatusCode)
	}

	b, err := io.ReadAll(r.Body)

	defer r.Body.Close()

	return b, r, err
}

func NewClient(ctx context.Context) MediaHostApi {
	return &mediaHostClient{}
}
