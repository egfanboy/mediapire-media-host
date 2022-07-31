package app

import (
	"sync"

	"github.com/egfanboy/mediapire-common/router"
)

type App struct {
	ControllerRegistry *router.ControllerRegistry
}

var a *App

var o = sync.Once{}

func initApp() {
	o.Do(func() {
		if a == nil {
			a = &App{ControllerRegistry: router.NewControllerRegistry()}
		}
	})
}

func GetApp() *App {
	initApp()

	return a
}

func init() {
	initApp()
}
