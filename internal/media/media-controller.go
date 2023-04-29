package media

import (
	"net/http"
	"strings"

	"github.com/egfanboy/mediapire-media-host/internal/app"
	"github.com/google/uuid"

	"github.com/egfanboy/mediapire-common/router"
)

const basePath = "/media"

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
		AddQueryParam(router.QueryParam{Name: "mediaType", Required: false}).
		SetPath(basePath).
		SetReturnCode(http.StatusOK).
		SetHandler(func(request *http.Request, p router.RouteParams) (interface{}, error) {
			mediaTypes := make([]string, 0)

			if p.Params["mediaType"] != "" {
				mediaTypes = strings.Split(p.Params["mediaType"], ",")
			}

			items, err := c.service.GetMedia(request.Context(), mediaTypes)
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
			return c.service.StreamMedia(request.Context(), uuid.MustParse(id))
		})
}

func initController() mediaController {
	c := mediaController{service: NewMediaService()}

	c.builders = append(c.builders, c.GetMedia, c.StreamMedia)

	return c
}

func init() {
	app.GetApp().ControllerRegistry.Register(initController())
}
