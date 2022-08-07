package app

import (
	"sync"

	"github.com/egfanboy/mediapire-common/router"
)

type App struct {
	ControllerRegistry *router.ControllerRegistry

	Config
}

var a *App

var o = sync.Once{}

func initApp() error {
	o.Do(func() {
		if a == nil {
			config, err := readConfig()

			if err != nil {
				return
			}
			a = &App{ControllerRegistry: router.NewControllerRegistry(), Config: config}
		}
	})

	return nil
}

func GetApp() *App {
	initApp()

	return a
}

func init() {
	// todo: handle error
	initApp()
}
