//go:build !wasm
// +build !wasm

package main

import (
	//"encoding/json"
	//"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	//"strconv"
	"syscall"
	//"time"

	"go.savla.dev/littr/backend"
	"go.savla.dev/littr/config"
	"go.savla.dev/littr/frontend"
	//"go.savla.dev/littr/models"

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
	app.Route("/tos", &frontend.ToSPage{})
	app.Route("/users", &frontend.UsersPage{})

	app.RunWhenOnBrowser()
}

func initServer() {
	// parse ENV contants from .env file (should be loaded using Makefile and docker-compose.yml file)
	config.ParseEnv()

	// create a channel for logging engine
	//models.LogsChan = make(chan models.Log, 1)

	// logging goroutine
	/*go func() {
		lg := <- models.LogsChan

		jsonData, err := json.Marshal(lg)
		if err != nil {
			log.Println(err.Error())
			return
		}

		fmt.Println(string(jsonData))
	}()*/

	// initialize caches
	backend.FlowCache = &core.Cache{}
	backend.PollCache = &core.Cache{}
	backend.SessionCache = &core.Cache{}
	backend.UserCache = &core.Cache{}

	log.Println("caches initialized")

	// load up data from local dumps (/opt/data/)
	backend.LoadData()

	log.Println("dumped data loaded")

	// run migrations
	ok := backend.RunMigrations()

	log.Println("migrations result:", ok)

	// handle system calls, signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// signals goroutine
	go func() {
		sig := <-sigs
		log.Printf("caught signal '%s', dumping data...", sig)

		backend.DumpData()
	}()

	// API routes
	http.HandleFunc("/api/auth", backend.AuthHandler)
	http.HandleFunc("/api/dump", backend.DumpHandler)
	http.HandleFunc("/api/flow", backend.FlowHandler)
	http.HandleFunc("/api/polls", backend.PollsHandler)
	http.HandleFunc("/api/stats", backend.StatsHandler)
	http.HandleFunc("/api/users", backend.UsersHandler)

	log.Println("API routes loaded")

	// root route with custom CSS and JS definitions
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
			"https://cdn.gscloud.cz/css/beer.min.css",
			//"/web/sortable.min.css",
		},
		Scripts: []string{
			"https://cdn.gscloud.cz/js/jquery.min.js",
			"https://cdn.gscloud.cz/js/beer.nomodule.min.js",
			"https://cdn.gscloud.cz/js/material-dynamic-colors.nomodule.min.js",
			//"https://cdn.gscloud.cz/js/litter.js?nonce=" + strconv.FormatInt(time.Now().UnixNano(), 10),
			"/web/litter.js",
		},
	})

	log.Println("starting the server...")

	// run a HTTP server
	if err := http.ListenAndServe(":8080", nil); err != nil {
		// https://github.com/maxence-charriere/go-app-demo/blob/V7/hello-docker/main.go
		//log.Fatal(err)
		panic(err)
	}

}
