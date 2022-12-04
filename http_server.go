//go:build !wasm
// +build !wasm

package main

import (
	"log"
	"net/http"

	"litter-go/backend"
	"litter-go/pages"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func initBackend() {}

func initWASM() {
	app.Route("/", &pages.LoginPage{})
	app.Route("/flow", &pages.FlowPage{})
	app.Route("/login", &pages.LoginPage{})
	app.Route("/logout", &pages.LoginPage{})
	app.Route("/polls", &pages.PollsPage{})
	app.Route("/post", &pages.PostPage{})
	app.Route("/register", &pages.RegisterPage{})
	app.Route("/settings", &pages.SettingsPage{})
	app.Route("/stats", &pages.StatsPage{})
	app.Route("/users", &pages.UsersPage{})

	app.RunWhenOnBrowser()
}

func initServer() {
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
			"https://cdn.jsdelivr.net/npm/beercss@2.3.0/dist/cdn/beer.min.css",
		},
		Scripts: []string{
			"https://cdn.jsdelivr.net/npm/beercss@2.3.0/dist/cdn/beer.min.js",
			"https://cdn.jsdelivr.net/npm/material-dynamic-colors@0.0.10/dist/cdn/material-dynamic-colors.min.js",
		},
	})

	http.HandleFunc("/api", backend.ApiHandleFunc)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

}
