package settings

import (
	"net/http"

	"github.com/egfanboy/mediapire-media-host/internal/app"

	"github.com/egfanboy/mediapire-common/router"
)

const basePath = "/settings"

type settingsController struct {
	builders []func() router.RouteBuilder
	service  settingsApi
}

func (c settingsController) GetApis() (routes []router.RouteBuilder) {
	for _, b := range c.builders {

		routes = append(routes, b())
	}

	return
}

func (c settingsController) HandleRegister() router.RouteBuilder {
	return router.NewV1RouteBuilder().
		SetMethod(http.MethodOptions, http.MethodGet).
		SetPath(basePath).
		SetReturnCode(http.StatusOK).
		SetHandler(func(request *http.Request, p router.RouteParams) (interface{}, error) {
			return c.service.GetSettings(request.Context())
		})
}

func initController() settingsController {
	c := settingsController{service: newSettingsService()}

	c.builders = append(c.builders, c.HandleRegister)

	return c
}

func init() {
	app.GetApp().ControllerRegistry.Register(initController())
}
