//go:build !wasm
// +build !wasm

package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.savla.dev/littr/backend"
	"go.savla.dev/littr/config"
	"go.savla.dev/littr/frontend"
	"go.savla.dev/swis/v5/pkg/core"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func initClient() {
	app.Route("/", &frontend.LoginPage{})
	app.Route("/flow", &frontend.FlowPage{})
	app.RouteWithRegexp("/flow/\\d+", &frontend.FlowPage{})
	app.RouteWithRegexp("/flow/\\w+", &frontend.FlowPage{})
	app.Route("/login", &frontend.LoginPage{})
	app.Route("/logout", &frontend.LoginPage{})
	app.Route("/polls", &frontend.PollsPage{})
	app.Route("/post", &frontend.PostPage{})
	app.Route("/register", &frontend.RegisterPage{})
	app.Route("/settings", &frontend.SettingsPage{})
	app.Route("/stats", &frontend.StatsPage{})
	app.Route("/tos", &frontend.ToSPage{})
	app.Route("/users", &frontend.UsersPage{})

	app.RunWhenOnBrowser()
}

func initServer() {
	// prepare the Logger instance
	l := backend.Logger{
		CallerID:   "system",
		WorkerName: "initServer",
	}

	// parse ENV contants from .env file (should be loaded using Makefile and docker-compose.yml file)
	config.ParseEnv()

	// initialize caches
	backend.FlowCache = &core.Cache{}
	backend.PollCache = &core.Cache{}
	backend.SubscriptionCache = &core.Cache{}
	backend.UserCache = &core.Cache{}

	l.Println("caches initialized", http.StatusOK)

	// load up data from local dumps (/opt/data/)
	// TODO: catch an error there!
	backend.LoadAll()

	l.Println("dumped data loaded", http.StatusOK)

	// run migrations
	backend.RunMigrations()

	// handle system calls, signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// signals goroutine
	go func() {
		sig := <-sigs
		l.Println("caught signal '"+sig.String()+"', dumping data...", http.StatusCreated)

		backend.DumpAll()
	}()

	// API routes
	http.HandleFunc("/api/auth", backend.AuthHandler)
	http.HandleFunc("/api/dump", backend.DumpHandler)
	http.HandleFunc("/api/flow", backend.FlowHandler)
	http.HandleFunc("/api/pix", backend.PixHandler)
	http.HandleFunc("/api/polls", backend.PollsHandler)
	http.HandleFunc("/api/push", backend.PushNotifHandler)
	//http.HandleFunc("/api/stats", backend.StatsHandler)
	http.HandleFunc("/api/users", backend.UsersHandler)

	l.Println("API routes loaded", http.StatusOK)

	// root route with custom CSS and JS definitions
	http.Handle("/", &app.Handler{
		Name:         "litter-go",
		ShortName:    "littr",
		Title:        "littr",
		Description:  "litter-go PWA",
		Author:       "krusty",
		LoadingLabel: "loading...",
		Lang:         "en",
		Icon: app.Icon{
			Large:      "/web/android-chrome-512x512.png",
			Default:    "/web/android-chrome-192x192.png",
			AppleTouch: "/web/apple-touch-icon.png",
		},
		Body: func() app.HTMLBody {
			return app.Body().Class("dark")
		},
		BackgroundColor: "#000000",
		ThemeColor:      "#000000",
		Version:         os.Getenv("APP_VERSION") + time.Now().String(),
		/*Preconnect: []string{
			"https://cdn.gscloud.cz/",
		},*/
		Styles: []string{
			"https://cdn.gscloud.cz/css/beer.min.css",
			"https://cdn.gscloud.cz/css//sortable.min.css",
			//"/web/sortable.min.css",
		},
		Scripts: []string{
			"https://cdn.gscloud.cz/js/jquery.min.js",
			"https://cdn.gscloud.cz/js/beer.nomodule.min.js",
			"https://cdn.gscloud.cz/js/material-dynamic-colors.nomodule.min.js",
			"https://cdn.gscloud.cz/js/sortable.min.js",
			"/web/litter.js",
		},
	})

	l.Println("starting the server...", http.StatusOK)

	// run a HTTP server
	if err := http.ListenAndServe(":8080", nil); err != nil {
		// https://github.com/maxence-charriere/go-app-demo/blob/V7/hello-docker/main.go
		//log.Fatal(err)
		panic(err)
	}

}
