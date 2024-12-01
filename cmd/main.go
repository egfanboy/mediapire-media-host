package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/egfanboy/mediapire-media-host/internal/app"
	"github.com/egfanboy/mediapire-media-host/internal/consul"
	"github.com/egfanboy/mediapire-media-host/internal/fs"
	_ "github.com/egfanboy/mediapire-media-host/internal/health"
	"github.com/egfanboy/mediapire-media-host/internal/media"
	"github.com/egfanboy/mediapire-media-host/internal/rabbitmq"
	_ "github.com/egfanboy/mediapire-media-host/internal/transfers"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	initiliazeApp()
}

var cleanupFuncs []func()

func addCleanupFunc(fn func()) {
	cleanupFuncs = append(cleanupFuncs, fn)
}

func initiliazeApp() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	log.Info().Msg("Initializing app")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		log.Info().Msg("Running cleanup functions")
		for _, fn := range cleanupFuncs {
			fn()
		}
	}()

	ctx := context.Background()

	err := rabbitmq.Setup(ctx)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to connect to rabbitmq")
		os.Exit(1)
	}

	addCleanupFunc(func() { rabbitmq.Cleanup() })
	fsService, _ := fs.NewFsService()
	mediaHost := app.GetApp()

	log.Debug().Msg("Setting up file system watchers")
	for _, d := range mediaHost.Directories {
		err := fsService.WatchDirectory(d)

		log.Error().Err(err).Str("Directory", d)
	}
	log.Debug().Msg("Finished setting up file system watchers")

	addCleanupFunc(fsService.CloseWatchers)

	// Scan the initial set of media
	log.Debug().Msg("Scanning media")
	mediaService := media.NewMediaService()

	err = mediaService.ScanDirectories(mediaHost.Directories...)
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

	addCleanupFunc(func() { srv.Close() })

	log.Debug().Msg("Registring ourselves to consul")

	err = consul.NewConsulClient()

	if err != nil {
		log.Error().Err(err).Msg("Failed to create the consul client")
	}

	err = consul.RegisterService()

	addCleanupFunc(func() { consul.UnregisterService() })

	if err != nil {
		log.Error().Err(err).Msg("Failed to register to consul")
	}

	log.Debug().Msg("Registration successful")

	log.Info().Msg("App initialized")

	<-c
}
