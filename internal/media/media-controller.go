package media

import (
	"net/http"
	"strings"

	"github.com/egfanboy/mediapire-media-host/internal/app"

	"github.com/egfanboy/mediapire-common/router"
)

const (
	basePath = "/media"
	pathArt  = "/media/{mediaId}/art"
)

var (
	queryParamOptionalMediaType = router.QueryParam{Name: "mediaType", Required: false}
	queryParamMediaId           = router.QueryParam{Name: "mediaId", Required: true}
	queryParamReturnContent     = router.QueryParam{Name: "returnContent", Required: true}
)

type mediaController struct {
	builders []func() router.RouteBuilder
	service  MediaApi
}

func (c mediaController) GetApis() (routes []router.RouteBuilder) {
	for _, b := range c.builders {

		routes = append(routes, b())
	}

	return
}

func (c mediaController) GetMedia() router.RouteBuilder {
	return router.NewV1RouteBuilder().
		SetMethod(http.MethodOptions, http.MethodGet).
		AddQueryParam(queryParamOptionalMediaType).
		SetPath(basePath).
		SetReturnCode(http.StatusOK).
		SetHandler(func(request *http.Request, p router.RouteParams) (interface{}, error) {
			mediaTypes := make([]string, 0)

			mediaTypeParam := p.Params[queryParamOptionalMediaType.Name]
			if mediaTypeParam != "" {
				mediaTypes = strings.Split(mediaTypeParam, ",")
			}

			items, err := c.service.GetMedia(request.Context(), mediaTypes)
			return items, err
		})
}

func (c mediaController) GetMediaById() router.RouteBuilder {
	return router.NewV1RouteBuilder().
		SetMethod(http.MethodOptions, http.MethodGet).
		AddQueryParam(queryParamMediaId).
		AddQueryParam(queryParamReturnContent).
		SetPath(basePath).
		SetReturnCode(http.StatusOK).
		SetHandler(func(request *http.Request, p router.RouteParams) (interface{}, error) {
			mediaId := router.MustGetQueryValue(p, queryParamMediaId)
			returnContent := router.MustGetQueryValue(p, queryParamReturnContent)

			// underlying router gives params as strings
			if returnContent == "true" {
				return c.service.GetMediaItemByIdWithContent(request.Context(), mediaId)
			} else {
				return c.service.GetMediaItemById(request.Context(), mediaId)
			}
		})
}

func (c mediaController) GetMediaArt() router.RouteBuilder {
	return router.NewV1RouteBuilder().
		SetMethod(http.MethodOptions, http.MethodGet).
		SetPath(pathArt).
		SetReturnCode(http.StatusOK).
		SetDataType(router.DataTypeFile).
		SetHandler(func(request *http.Request, p router.RouteParams) (interface{}, error) {
			mediaId := p.Params["mediaId"]

			items, err := c.service.GetMediaArt(request.Context(), mediaId)
			return items, err
		})
}

func (c mediaController) StreamMedia() router.RouteBuilder {
	return router.NewV1RouteBuilder().
		SetMethod(http.MethodOptions, http.MethodGet).
		SetPath(basePath + "/stream").
		SetReturnCode(200).
		SetDataType(router.DataTypeFile).
		AddQueryParam(router.QueryParam{Name: "id", Required: true}).
		SetHandler(func(request *http.Request, p router.RouteParams) (interface{}, error) {
			id := p.Params["id"]
			return c.service.StreamMedia(request.Context(), id)
		})
}

func (c mediaController) DownloadMedia() router.RouteBuilder {
	return router.NewV1RouteBuilder().
		SetMethod(http.MethodOptions, http.MethodPost).
		SetPath(basePath + "/download").
		SetReturnCode(200).
		SetDataType(router.DataTypeFile).
		SetHandler(func(request *http.Request, p router.RouteParams) (interface{}, error) {
			var body []string
			err := p.PopulateBody(&body)
			if err != nil {
				return nil, err
			}
			return c.service.DownloadMedia(request.Context(), body)
		})
}

func initController() mediaController {
	c := mediaController{service: NewMediaService()}

	c.builders = append(c.builders,
		c.GetMedia,
		c.StreamMedia,
		c.DownloadMedia,
		c.GetMediaArt,
		c.GetMediaById,
	)

	return c
}

func init() {
	app.GetApp().ControllerRegistry.Register(initController())
}
