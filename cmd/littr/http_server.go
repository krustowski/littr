//go:build !wasm
// +build !wasm

package main

import (
	"compress/flate"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	be "go.savla.dev/littr/pkg/backend"
	"go.savla.dev/littr/pkg/backend/common"
	"go.savla.dev/littr/pkg/backend/db"
	"go.savla.dev/littr/pkg/backend/posts"
	fe "go.savla.dev/littr/pkg/frontend"
	"go.savla.dev/swis/v5/pkg/core"

	sse "github.com/alexandrevicenzi/go-sse"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func initClient() {
	app.Route("/", &fe.LoginPage{})
	app.Route("/flow", &fe.FlowPage{})
	app.RouteWithRegexp("/flow/post/\\d+", &fe.FlowPage{})
	app.RouteWithRegexp("/flow/hashtag/\\w+", &fe.FlowPage{})
	app.RouteWithRegexp("/flow/user/\\w+", &fe.FlowPage{})
	app.Route("/login", &fe.LoginPage{})
	app.Route("/logout", &fe.LoginPage{})
	app.Route("/polls", &fe.PollsPage{})
	app.Route("/post", &fe.PostPage{})
	app.Route("/register", &fe.RegisterPage{})
	app.Route("/reset", &fe.ResetPage{})
	app.Route("/settings", &fe.SettingsPage{})
	app.Route("/stats", &fe.StatsPage{})
	app.Route("/tos", &fe.ToSPage{})
	app.Route("/users", &fe.UsersPage{})

	app.RunWhenOnBrowser()
}

func initServer() {
	r := chi.NewRouter()

	r.Use(middleware.CleanPath)
	// TODO
	//r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	compressor := middleware.NewCompressor(
		//flate.DefaultCompression,
		flate.BestCompression,
		"application/wasm", "text/css", "image/svg+xml", "application/json", "image/gif",
	)
	r.Use(compressor.Handler)

	// custom listener
	// https://github.com/oderwat/go-nats-app/blob/master/back/back.go
	listener, err := net.Listen("tcp", ":"+os.Getenv("APP_PORT"))
	if err != nil {
		panic(err)
	}

	// custom server
	// https://github.com/oderwat/go-nats-app/blob/master/back/back.go
	server := &http.Server{
		Addr: listener.Addr().String(),
		//ReadTimeout: 0 * time.Second,
		WriteTimeout: 0 * time.Second,
	}

	// prepare the Logger instance
	l := common.Logger{
		CallerID:   "system",
		WorkerName: "initServer",
		Version:    "system",
	}

	// initialize caches
	db.FlowCache = &core.Cache{}
	db.PollCache = &core.Cache{}
	db.SubscriptionCache = &core.Cache{}
	db.TokenCache = &core.Cache{}
	db.UserCache = &core.Cache{}

	l.Println("caches initialized", http.StatusOK)

	// load up data from local dumps (/opt/data/)
	// TODO: catch an error there!
	db.LoadAll()

	l.Println("dumped data loaded", http.StatusOK)

	// run migrations
	db.RunMigrations()

	// handle system calls, signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// signals goroutine
	go func() {
		sig := <-sigs
		signal.Stop(sigs)

		l.Println("caught signal '"+sig.String()+"', exiting gracefully...", http.StatusCreated)
		if posts.Streamer != nil {
			posts.Streamer.SendMessage("/api/v1/posts/live", sse.SimpleMessage("server-stop"))
		}

		// TODO
		//db.LockAll()
		db.DumpAll()
		posts.Streamer.Shutdown()
	}()

	// parse the custom Service Worker template string for the app handler
	tpl, err := os.ReadFile("/opt/web/app-worker.js")
	if err != nil {
		panic(err)
	}

	swTemplateString := fmt.Sprintf("%s", tpl)

	// API router
	r.Mount("/api/v1", be.APIRouter())

	appHandler := &app.Handler{
		Name:         "litter-go",
		ShortName:    "littr",
		Title:        "littr",
		Description:  "A simple nanoblogging platform",
		Author:       "krusty",
		LoadingLabel: "loading...",
		Lang:         "en",
		Keywords: []string{
			"go-app",
			"nanoblogging",
			"nanoblog",
			"microblogging",
			"microblog",
			"blogging",
			"blog",
			"social network",
		},
		AutoUpdateInterval: time.Minute * 1,
		Icon: app.Icon{
			Default:    "/web/android-chrome-192x192.png",
			SVG:        "/web/android-chrome-512x512.svg",
			Large:      "/web/android-chrome-512x512.png",
			AppleTouch: "/web/apple-touch-icon.png",
		},
		Image: "/web/android-chrome-512x512.svg",
		Body: func() app.HTMLBody {
			return app.Body().Class("dark")
		},
		BackgroundColor: "#000000",
		ThemeColor:      "#000000",
		Version:         os.Getenv("APP_VERSION") + "-" + time.Now().Format("2006-01-02_15:04:05"),
		Env: map[string]string{
			"APP_VERSION":          os.Getenv("APP_VERSION"),
			"APP_ENVIRONMENT":      os.Getenv("APP_ENVIRONMENT"),
			"REGISTRATION_ENABLED": os.Getenv("REGISTRATION_ENABLED"),
			"VAPID_PUB_KEY":        os.Getenv("VAPID_PUB_KEY"),
		},
		Preconnect: []string{
			//"https://cdn.savla.dev/",
		},
		Fonts: []string{
			"https://cdn.savla.dev/css/material-symbols-outlined.woff2",
			//"https://cdn.jsdelivr.net/npm/beercss@3.5.0/dist/cdn/material-symbols-outlined.woff2",
		},
		Styles: []string{
			"https://cdn.savla.dev/css/beercss.min.css",
			"https://cdn.savla.dev/css/sortable.min.css",
			"/web/litter.css",
		},
		Scripts: []string{
			"https://cdn.savla.dev/js/jquery.min.js",
			"https://cdn.savla.dev/js/beer.nomodule.min.js",
			"https://cdn.savla.dev/js/material-dynamic-colors.nomodule.min.js",
			"https://cdn.savla.dev/js/sortable.min.js",
			"/web/litter.js",
			//"https://cdn.savla.dev/js/litter.js",
			"https://cdn.savla.dev/js/eventsource.min.js",
		},
		ServiceWorkerTemplate: swTemplateString,
	}

	r.Handle("/*", appHandler)
	server.Handler = r

	l.Println("starting the server...", http.StatusOK)
	go serverStartNotif()

	// TODO use http.ErrServerClosed for graceful shutdown
	if err := server.Serve(listener); err != nil {
		panic(err)
	}
}

func serverStartNotif() {
	time.Sleep(time.Second * 30)

	if posts.Streamer != nil {
		posts.Streamer.SendMessage("/api/v1/posts/live", sse.SimpleMessage("server-start"))
	}

}
