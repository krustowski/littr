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

	be "go.vxn.dev/littr/pkg/backend"
	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/backend/db"
	"go.vxn.dev/littr/pkg/backend/posts"
	fe "go.vxn.dev/littr/pkg/frontend"
	"go.vxn.dev/swis/v5/pkg/core"

	sse "github.com/alexandrevicenzi/go-sse"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func initClient() {
	app.Route("/", &fe.WelcomeView{})
	app.Route("/flow", &fe.FlowView{})
	app.RouteWithRegexp("/flow/post/\\d+", &fe.FlowView{})
	app.RouteWithRegexp("/flow/hashtag/\\w+", &fe.FlowView{})
	app.RouteWithRegexp("/flow/user/\\w+", &fe.FlowView{})
	app.Route("/login", &fe.LoginView{})
	app.Route("/logout", &fe.LoginView{})
	app.Route("/polls", &fe.PollsView{})
	app.Route("/post", &fe.PostView{})
	app.Route("/register", &fe.RegisterView{})
	app.Route("/reset", &fe.ResetView{})
	app.RouteWithRegexp("/reset/\\w+", &fe.ResetView{})
	app.Route("/settings", &fe.SettingsView{})
	app.Route("/stats", &fe.StatsView{})
	app.Route("/tos", &fe.ToSView{})
	app.Route("/users", &fe.UsersView{})

	app.RunWhenOnBrowser()
}

type Handler func(w http.ResponseWriter, r *http.Request) error

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {
		// handle returned error here.
		w.WriteHeader(500)
		w.Write([]byte("empty"))
	}
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
		"application/wasm", "text/css", "image/svg+xml", "application/json", "image/gif", "application/octet-stream",
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
	db.RequestCache = &core.Cache{}
	db.SubscriptionCache = &core.Cache{}
	db.TokenCache = &core.Cache{}
	db.UserCache = &core.Cache{}

	l.Println("caches initialized", http.StatusOK)

	// load up data from local dumps (/opt/data/)
	// TODO: catch an error there!
	loadReport := db.LoadAll()
	l.Println(loadReport, http.StatusOK)

	l.Println("dumped data loaded", http.StatusOK)

	// run migrations
	db.RunMigrations()

	// handle system calls, signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// signals goroutine
	go func() {
		// wait for signals
		sig := <-sigs
		signal.Stop(sigs)

		l.Msg("trap signal: " + sig.String() + ", exiting gracefully...").Status(http.StatusOK)
		if posts.Streamer != nil {
			posts.Streamer.SendMessage("/api/v1/posts/live", sse.SimpleMessage("server-stop"))
		}

		// TODO
		//db.LockAll()
		l.Msg(db.DumpAll()).Status(http.StatusOK).Log()

		if posts.Streamer != nil {
			posts.Streamer.Shutdown()
		}
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
		Name:         "littr",
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
			"platform",
			"blogging",
			"board",
			"blog",
			"social network",
		},
		AutoUpdateInterval: time.Minute * 1,
		Icon: app.Icon{
			//Maskable:   "/web/android-chrome-192x192.png",
			Default:    "/web/android-chrome-192x192.png",
			SVG:        "/web/android-chrome-512x512.svg",
			Large:      "/web/android-chrome-512x512.png",
			AppleTouch: "/web/apple-touch-icon.png",
		},
		Image: "/web/android-chrome-512x512.svg",
		//Domain: "www.littr.eu",
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
			//"https://cdn.vxn.dev/",
		},
		Fonts: []string{
			"https://cdn.vxn.dev/css/material-symbols-outlined.woff2",
			//"https://cdn.jsdelivr.net/npm/beercss@3.5.0/dist/cdn/material-symbols-outlined.woff2",
		},
		Styles: []string{
			"https://cdn.vxn.dev/css/beercss-3.7.0-custom.min.css",
			"https://cdn.vxn.dev/css/sortable.min.css",
			"/web/littr.css",
		},
		Scripts: []string{
			"https://cdn.vxn.dev/js/jquery.min.js",
			"https://cdn.vxn.dev/js/beer.min.js",
			//"https://cdn.jsdelivr.net/npm/beercss@3.7.0/dist/cdn/beer.min.js",
			"https://cdn.vxn.dev/js/material-dynamic-colors.min.js",
			//"https://cdn.jsdelivr.net/npm/material-dynamic-colors@1.1.2/dist/cdn/material-dynamic-colors.min.js",
			"https://cdn.vxn.dev/js/sortable.min.js",
			"https://cdn.vxn.dev/js/eventsource.min.js",
			"/web/littr.js",
			//"https://cdn.vxn.dev/js/littr.js",
		},
		ServiceWorkerTemplate: swTemplateString,
	}

	r.Method("GET", "/favicon.ico", Handler(func(w http.ResponseWriter, r *http.Request) error {
		http.ServeFile(w, r, "/opt/web/favicon.ico")
		return nil
	}))

	r.Handle("/*", appHandler)
	server.Handler = r

	l.Println("starting the server (v"+os.Getenv("APP_VERSION")+")...", http.StatusOK)
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
