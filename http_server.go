//go:build !wasm
// +build !wasm

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"litter-go/backend"
	"litter-go/config"
	"litter-go/frontend"
	"litter-go/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"go.savla.dev/swis/v5/pkg/core"
)

func initClient() {
	app.Route("/", &frontend.LoginPage{})
	app.Route("/flow", &frontend.FlowPage{})
	app.Route("/login", &frontend.LoginPage{})
	app.Route("/logout", &frontend.LoginPage{})
	app.Route("/polls", &frontend.PollsPage{})
	app.Route("/post", &frontend.PostPage{})
	app.Route("/register", &frontend.RegisterPage{})
	app.Route("/settings", &frontend.SettingsPage{})
	app.Route("/stats", &frontend.StatsPage{})
	app.Route("/users", &frontend.UsersPage{})

	app.RunWhenOnBrowser()
}

func initServer() {
	config.ParseEnv()

	// create a channel for logging engine
	config.LogsChan := make(chan models.Log, 1)

	// logging goroutine
	go func() {
		lg := <- config.LogsChan

		jsonData, err := json.Marshal(lg)
		if err != nil {
			log.Println(err.Error())
			return err
		}

		log.Println(jsonData)
	}()

	// initialize caches
	backend.FlowCache = &core.Cache{}
	backend.PollCache = &core.Cache{}
	backend.SessionCache = &core.Cache{}
	backend.UserCache = &core.Cache{}

	// load up data from local dumps (/opt/data/)
	backend.LoadData()

	// handle system calls, signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// signals goroutine
	go func() {
		sig := <-sigs
		log.Printf("caught signal '%s', dumping data...", sig)

		backend.DumpData()
	}()

	http.HandleFunc("/api/auth", backend.AuthHandler)
	http.HandleFunc("/api/flow", backend.FlowHandler)
	http.HandleFunc("/api/stats", backend.StatsHandler)
	http.HandleFunc("/api/users", backend.UsersHandler)

	http.Handle("/", &app.Handler{
		Name:        "litter-go",
		Description: "litter-go PWA",
		Author:      "krusty",
		Icon: app.Icon{
			Default:    "/web/android-chrome-512x512.png",
			AppleTouch: "/web/apple-touch-icon.png",
		},
		BackgroundColor: "#000000",
		ThemeColor:      "#000000",
		Styles: []string{
			"https://cdn.jsdelivr.net/npm/beercss@3.2.13/dist/cdn/beer.min.css",
		},
		Scripts: []string{
			"https://cdn.jsdelivr.net/npm/beercss@3.2.13/dist/cdn/beer.min.js",
			"https://cdn.jsdelivr.net/npm/material-dynamic-colors@1.0.1/dist/cdn/material-dynamic-colors.min.js",
		},
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

}
