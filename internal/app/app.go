package app

import (
	"os"
	"sync"

	"github.com/egfanboy/mediapire-common/router"
	"github.com/google/uuid"
)

type App struct {
	ControllerRegistry *router.ControllerRegistry

	config
	NodeId uuid.UUID
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

func initApp() {
	if a == nil {
		config, err := readConfig()

		if err != nil {
			return
		}
		a = &App{ControllerRegistry: router.NewControllerRegistry(), config: config}
	}

	// Create the download path from the config in case it does not exist
	err := os.MkdirAll(a.config.DownloadPath, os.ModePerm)
	if err != nil {
		return
	}
}

func createApp() error {
	o.Do(initApp)

	return nil
}

func GetApp() *App {
	createApp()

	return a
}

func init() {
	// todo: handle error
	initApp()
}
