package app

import (
	"sync"

	"github.com/egfanboy/mediapire-common/router"
)

type App struct {
	ControllerRegistry *router.ControllerRegistry

	config
}

func (a *App) IsMediaSupported(extension string) bool {

	for _, t := range a.config.FileTypes {
		if extension == t {
			return true
		}
	}

	return false
}

func (a *App) GetDirectories() []string {
	return a.Directories
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
			a = &App{ControllerRegistry: router.NewControllerRegistry(), config: config}
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
