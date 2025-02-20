package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/egfanboy/mediapire-media-host/pkg/types"
)

const (
	baseMediaPath     = "/api/v1/media"
	baseTransfersPath = "/api/v1/transfers"
	baseSettingsPath  = "/api/v1/settings"
)

type MediaHostApi interface {
	GetMedia(ctx context.Context, mediaTypes *[]string) ([]types.MediaItem, *http.Response, error)
	StreamMedia(ctx context.Context, mediaId string) ([]byte, *http.Response, error)
	DownloadTransfer(ctx context.Context, transferId string) ([]byte, *http.Response, error)
	GetSettings(ctx context.Context) (types.MediaHostSettings, *http.Response, error)
	GetMediaArt(ctx context.Context, mediaId string) ([]byte, *http.Response, error)
}

func buildUriFromHost(h types.Host, apiUri string) string {
	return fmt.Sprintf("%s://%s:%v%s", h.Scheme(), h.Host(), h.Port(), apiUri)
}

type mediaHostClient struct {
	host types.Host
}

func (c *mediaHostClient) GetMedia(ctx context.Context, mediaTypes *[]string) (result []types.MediaItem, r *http.Response, err error) {
	apiUrl := baseMediaPath

	if mediaTypes != nil && len(*mediaTypes) > 0 {
		apiUrl = apiUrl + fmt.Sprintf("?mediaType=%s", strings.Join(*mediaTypes, ","))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, buildUriFromHost(c.host, apiUrl), nil)
	if err != nil {
		return
	}

	r, err = http.DefaultClient.Do(req)
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

func (c *mediaHostClient) StreamMedia(ctx context.Context, mediaId string) (b []byte, r *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, buildUriFromHost(c.host, fmt.Sprintf("%s/stream?id=%s", baseMediaPath, mediaId)), nil)
	if err != nil {
		return
	}

	r, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	b, err = io.ReadAll(r.Body)

	defer r.Body.Close()

	return
}

func (c *mediaHostClient) DownloadTransfer(ctx context.Context, transferId string) ([]byte, *http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, buildUriFromHost(c.host, fmt.Sprintf("%s/%s/download", baseTransfersPath, transferId)), nil)
	if err != nil {
		return nil, nil, err
	}

	r, err := http.DefaultClient.Do(req)
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

func (c *mediaHostClient) GetSettings(ctx context.Context) (result types.MediaHostSettings, r *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, buildUriFromHost(c.host, baseSettingsPath), nil)
	if err != nil {
		return
	}

	r, err = http.DefaultClient.Do(req)
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

func (c *mediaHostClient) GetMediaArt(ctx context.Context, mediaId string) ([]byte, *http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, buildUriFromHost(c.host, fmt.Sprintf("%s/%s/art", baseMediaPath, mediaId)), nil)
	if err != nil {
		return nil, nil, err
	}

	r, err := http.DefaultClient.Do(req)
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

func NewClient(h types.Host) MediaHostApi {
	return &mediaHostClient{host: h}
}
