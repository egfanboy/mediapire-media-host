package main

import (
	"fmt"
	"mediapire-media-host/cmd/app"
	"mediapire-media-host/cmd/fs"
	_ "mediapire-media-host/cmd/health"
	"mediapire-media-host/cmd/media"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	initiliazeApp()
}

func initiliazeApp() {

	c := make(chan struct{})

	fsService, _ := fs.NewFsService()
	mediaHost := app.GetApp()

	for _, d := range mediaHost.Directories {
		fsService.WatchDirectory(d)
	}

	defer fsService.CloseWatchers()

	// Scan the initial set of media
	mediaService := media.NewMediaService()

	err := mediaService.ScanDirectories(mediaHost.Directories...)

	if err != nil {
		fmt.Print(err.Error())
	}

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

	fmt.Println("Starting server")

	srv.ListenAndServe()

	defer srv.Close()
	<-c
}
