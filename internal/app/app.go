package app

import (
	"os"
	"sync"

	"github.com/egfanboy/mediapire-common/router"
	"github.com/google/uuid"

	"github.com/rs/zerolog/log"
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
			log.Error().Err(err).Msg("Error occured when trying to instantiate the application. Exiting.")
			os.Exit(1)
			return
		}
		a = &App{ControllerRegistry: router.NewControllerRegistry(), config: config}
	}

	// Create the download path from the config in case it does not exist
	err := os.MkdirAll(a.config.DownloadPath, os.ModePerm)
	if err != nil {
		return
	}

	// Create the art path from the config in case it does not exist
	err = os.MkdirAll(a.config.ArtPath, os.ModePerm)
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
	initApp()
}
