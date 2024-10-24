//go:build !wasm
// +build !wasm

package main

import (
	"compress/flate"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	be "go.vxn.dev/littr/pkg/backend"
	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/backend/db"
	"go.vxn.dev/littr/pkg/backend/live"
	"go.vxn.dev/littr/pkg/backend/metrics"
	"go.vxn.dev/littr/pkg/config"
	fe "go.vxn.dev/littr/pkg/frontend"
	"go.vxn.dev/swis/v5/pkg/core"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Handler func(w http.ResponseWriter, r *http.Request) error

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {
		// handle returned error here.
		w.WriteHeader(500)
		w.Write([]byte("empty"))
	}
}

// appHandler hold the pointer to the very main FE app handler.
var appHandler = &app.Handler{
	Name:         "littr",
	ShortName:    "littr",
	Title:        "littr",
	Description:  "A simple nanoblogging platform",
	Author:       "krusty",
	LoadingLabel: "loading...",
	Lang:         "en",
	Keywords: []string{
		"blog",
		"blogging",
		"board",
		"go-app",
		"microblog",
		"microblogging",
		"nanoblog",
		"nanoblogging",
		"platform",
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
	ServiceWorkerTemplate: config.EnchartedSW,
}

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

var (
	server *http.Server
	wg     sync.WaitGroup
)

func initServer() {
	// Prepare the Logger instance.
	l := common.NewLogger(nil, "initServer")

	//
	//  Prestart preparations
	//

	// Handle system calls and signals to properly shutdown the server.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// The signals monitoring goroutine.
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Wait for signals.
		sig := <-sigs
		signal.Stop(sigs)

		// Log and broadcast the message that the server is to shutdown.
		l.Msg("trap signal: " + sig.String() + ", stopping the HTTP server gracefully...").Status(http.StatusOK).Log()
		live.BroadcastMessage("server-stop", "message")

		// Fetch a context to send to gracefully shutdown the HTTP server.
		sctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		// Lock the database so no one can change any data henceforth.
		db.Lock()

		// Dump all in-memory databases.
		l.Msg(db.DumpAll()).Status(http.StatusOK).Log()

		// Terminate the server from here.
		if err := server.Shutdown(sctx); err != nil {
			l.Msg(fmt.Sprintf("HTTP server shutdown error: %s, force terminitaing the server...", err.Error())).Status(http.StatusInternalServerError).Log()

			// Force terminate the server if failed to stop gracefully.
			server.Close()
			return
		}

		l.Msg("graceful shutdown complete").Status(http.StatusOK).Log()
	}()

	//
	//  Muxer, listener and server initialization
	//

	// Create a new go-chi muxer.
	r := chi.NewRouter()

	r.Use(middleware.CleanPath)
	//r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Enable a proactive data compression.
	compressor := middleware.NewCompressor(
		flate.BestCompression,
		"application/wasm", "text/css", "image/svg+xml", "application/json", "image/gif", "application/octet-stream",
	)
	r.Use(compressor.Handler)

	// Create a custom network connection listener.
	listener, err := net.Listen("tcp", ":"+config.ServerPort)
	if err != nil {
		// Cannot listen on such address = permission issue?
		panic(err)
	}
	defer listener.Close()

	// Create a custom HTTP server.
	server = &http.Server{
		Addr: listener.Addr().String(),
		//ReadTimeout: 0 * time.Second,
		WriteTimeout: 0 * time.Second,
		Handler:      r,
	}

	//
	//  Database and data initialization
	//

	// Initialize all the in-memory databases (caches).
	db.FlowCache = &core.Cache{Name: "FlowCache"}
	db.PollCache = &core.Cache{Name: "PollCache"}
	db.RequestCache = &core.Cache{Name: "RequestCache"}
	db.SubscriptionCache = &core.Cache{Name: "SubscriptionCache"}
	db.TokenCache = &core.Cache{Name: "TokenCache"}
	db.UserCache = &core.Cache{Name: "UserCache"}

	// Unlock the write access.
	db.Unlock()

	l.Msg("caches initialized").Status(http.StatusOK).Log()

	// Register the (mostly) cache metrics.
	metrics.RegisterAll()

	// Load the persistent data from the filesystem to memory.
	l.Msg(db.LoadAll()).Status(http.StatusOK).Log()

	l.Msg("dumped data loaded").Status(http.StatusOK).Log()

	// Run data migration procedures to the database schema.
	l.Msg(db.RunMigrations(l)).Status(http.StatusOK).Log()

	//
	//  Routes and handlers mounting
	//

	// Mount the very main API router spanning all the backend.
	r.Mount("/api/v1", be.APIRouter())

	// A workaround to serve a proper favicon icon.
	r.Method("GET", "/favicon.ico", Handler(func(w http.ResponseWriter, r *http.Request) error {
		http.ServeFile(w, r, "/opt/web/favicon.ico")
		return nil
	}))

	// Register the Prometheus metrics' handle.
	r.Handle("/metrics", promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{
		Registry: metrics.Registry,
	}))

	// Handle the rest using the appHandler defined earlier.
	r.Handle("/*", appHandler)

	//
	//  Start the server
	//

	l.Msg("starting the server (v" + os.Getenv("APP_VERSION") + ")...").Status(http.StatusOK).Log()

	// Send the SSE
	go func() {
		time.Sleep(time.Second * 30)
		live.BroadcastMessage("server-start", "message")
	}()

	// Start serving via the created net listener.
	if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		l.Msg(fmt.Sprintf("HTTP server error: %s", err.Error())).Status(http.StatusInternalServerError).Log()
	}

	//
	//  Exit
	//

	// Wait for the graceful HTTP server shutdown attempt.
	wg.Wait()

	// This is the final log before the application exits for real!
	// https://dev.to/mokiat/proper-http-shutdown-in-go-3fji
	l.Msg("HTTP server has stopped serving new connections, exit").Status(http.StatusOK).Log()
}
