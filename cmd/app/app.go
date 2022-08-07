package app

import (
	"log"
	"os"
	"sync"

	"github.com/egfanboy/mediapire-common/router"
)

type App struct {
	ControllerRegistry *router.ControllerRegistry
	Logger             *log.Logger
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
			a = &App{ControllerRegistry: router.NewControllerRegistry(), Logger: log.New(os.Stdout, "mediapire.media-host.main", log.Ldate|log.Ltime), Config: config}
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
