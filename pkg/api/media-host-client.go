package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/egfanboy/mediapire-media-host/pkg/types"
	"github.com/google/uuid"
)

const (
	baseMediaPath = "/api/v1/media"
)

type MediaHostApi interface {
	GetHealth(h types.Host) (*http.Response, error)
	GetMedia(h types.Host) ([]types.MediaItem, *http.Response, error)
	StreamMedia(h types.Host, mediaId uuid.UUID) ([]byte, *http.Response, error)
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

	body, err := ioutil.ReadAll(r.Body)
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

	b, err = ioutil.ReadAll(r.Body)

	defer r.Body.Close()

	return
}

func NewClient(ctx context.Context) MediaHostApi {
	return &mediaHostClient{}
}
