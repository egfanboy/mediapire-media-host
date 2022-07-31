package main

import (
	"fmt"
	"log"
	"mediapire-media-host/cmd/app"
	_ "mediapire-media-host/cmd/health"

	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

var (
	logger = log.New(os.Stdout, "mediapire.media-host.main", log.Ldate|log.Ltime)
)

func main() {
	fmt.Print("start server")
	mainRouter := mux.NewRouter()

	for _, c := range app.GetApp().ControllerRegistry.GetControllers() {
		for _, b := range c.GetApis() {
			b.Build(mainRouter)
		}
	}

	srv := &http.Server{
		Addr: "0.0.0.0:9797",
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      mainRouter, // Pass our instance of gorilla/mux in.
	}

	srv.ListenAndServe()
}
