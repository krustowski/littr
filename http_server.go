//go:build !wasm
// +build !wasm

package main

import (
	"compress/flate"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.savla.dev/littr/backend"
	"go.savla.dev/littr/config"
	"go.savla.dev/littr/frontend"
	"go.savla.dev/swis/v5/pkg/core"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
	r := chi.NewRouter()

	r.Use(middleware.CleanPath)
	//r.Use(middleware.Logger)
	compressor := middleware.NewCompressor(flate.DefaultCompression,
		"application/wasm", "text/css", "image/svg+xml", "application/json")
	r.Use(compressor.Handler)
	r.Use(middleware.Recoverer)

	// custom listener
	// https://github.com/oderwat/go-nats-app/blob/master/back/back.go
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	// custom server
	// https://github.com/oderwat/go-nats-app/blob/master/back/back.go
	server := &http.Server{
		Addr: listener.Addr().String(),
	}

	// prepare the Logger instance
	l := backend.Logger{
		CallerID:   "system",
		WorkerName: "initServer",
		Version:    "system",
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

	// API router
	r.Mount("/api", backend.LoadAPIRouter())

	appHandler := &app.Handler{
		Name:               "litter-go",
		ShortName:          "littr",
		Title:              "littr",
		Description:        "litter-go PWA",
		Author:             "krusty",
		LoadingLabel:       "loading...",
		Lang:               "en",
		AutoUpdateInterval: time.Minute * 5,
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
	}

	r.Handle("/*", appHandler)
	server.Handler = r

	l.Println("starting the server...", http.StatusOK)

	// TODO use http.ErrServerClosed for graceful shutdown
	if err := server.Serve(listener); err != nil {
		panic(err)
	}
}
