package main

import (
	"mediapire-media-host/cmd/app"
	"mediapire-media-host/cmd/fs"
	_ "mediapire-media-host/cmd/health"
	"mediapire-media-host/cmd/media"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	initiliazeApp()
}

func initiliazeApp() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	log.Info().Msg("Initializing app")
	c := make(chan struct{})

	fsService, _ := fs.NewFsService()
	mediaHost := app.GetApp()

	log.Debug().Msg("Setting up file system watchers")
	for _, d := range mediaHost.Directories {
		err := fsService.WatchDirectory(d)

		log.Error().Err(err).Str("Directory", d)
	}
	log.Debug().Msg("Finished setting up file system watchers")
	defer fsService.CloseWatchers()

	// Scan the initial set of media
	log.Debug().Msg("Scanning media")
	mediaService := media.NewMediaService()

	err := mediaService.ScanDirectories(mediaHost.Directories...)

	if err != nil {
		log.Error().Err(err)
	}

	log.Debug().Msg("Finished scanning media")

	log.Debug().Msg("Starting webserver")
	mainRouter := mux.NewRouter()

	for _, c := range app.GetApp().ControllerRegistry.GetControllers() {
		for _, b := range c.GetApis() {
			b.Build(mainRouter)
		}
	}

	srv := &http.Server{
		Addr:         "0.0.0.0:9797",
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      mainRouter,
	}

	srv.ListenAndServe()

	log.Debug().Msg("webserver started")

	defer srv.Close()
	<-c
}
