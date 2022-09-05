package media

import (
	"net/http"

	"github.com/egfanboy/mediapire-media-host/internal/app"

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
		SetPath(basePath).
		SetReturnCode(http.StatusOK).
		SetHandler(func(request *http.Request, p router.RouteParams) (interface{}, error) {
			items, err := c.service.GetMedia(request.Context())
			return items, err
		})
}

func (c mediaController) StreamMedia() router.RouteBuilder {
	return router.NewV1RouteBuilder().
		SetMethod(http.MethodOptions, http.MethodGet).
		SetPath(basePath + "/stream").
		SetReturnCode(200).
		SetDataType(router.DataTypeFile).
		AddQueryParam(router.QueryParam{Name: "filePath", Required: true}).
		SetHandler(func(request *http.Request, p router.RouteParams) (interface{}, error) {
			return c.service.StreamMedia(request.Context(), p.Params["filePath"])
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
