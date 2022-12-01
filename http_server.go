//go:build !wasm
// +build !wasm

package main

import (
	"log"
	"net/http"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func initWASM() {
	app.Route("/flow", &flowPage{})
	app.Route("/login", &loginPage{})
	app.Route("/polls", &pollsPage{})
	app.Route("/settings", &settingsPage{})
	app.Route("/stats", &statsPage{})
	app.Route("/users", &usersPage{})

	app.RunWhenOnBrowser()
}

func initServer() {
	http.Handle("/", &app.Handler{
		Name:        "litter-go",
		Description: "litter-go PWA",
		Icon: app.Icon{
			Default:    "/web/logo_284.png",
			AppleTouch: "/web/apple-touch-icon.png",
		},
		Styles: []string{
			"https://cdn.jsdelivr.net/npm/beercss@2.3.0/dist/cdn/beer.min.css",
		},
		Scripts: []string{
			"https://cdn.jsdelivr.net/npm/beercss@2.3.0/dist/cdn/beer.min.js",
			"https://cdn.jsdelivr.net/npm/material-dynamic-colors@0.0.10/dist/cdn/material-dynamic-colors.min.js",
		},
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
