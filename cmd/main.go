package main

import (
	"fmt"
	"mediapire-media-host/cmd/app"
	"mediapire-media-host/cmd/fs"
	_ "mediapire-media-host/cmd/health"
	"mediapire-media-host/cmd/integration/master"
	"mediapire-media-host/cmd/media"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	initiliazeApp()
}

func initiliazeApp() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

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
		Addr:         fmt.Sprintf("0.0.0.0:%d", mediaHost.Port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      mainRouter,
	}

	go func() error {
		return srv.ListenAndServe()
	}()

	log.Debug().Msg("Webserver started")

	defer srv.Close()

	log.Debug().Msg("Calling master node to register ourselves")
	err = master.NewMasterIntegration().RegisterNode(mediaHost.Master.Scheme, mediaHost.Master.Host, mediaHost.Master.Port)

	if err != nil {
		log.Error().Err(err).Msg("Failed to register to master node, exiting")
		os.Exit(1)
	}

	log.Debug().Msg("Registration successful")

	log.Info().Msg("App initialized")

	<-c
}
