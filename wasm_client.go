//go:build wasm
// +build wasm

package main

import "github.com/maxence-charriere/go-app/v9/pkg/app"

func initWASM() {
	app.Route("/flow", &flowPage{})
	app.Route("/login", &loginPage{})
	app.Route("/polls", &pollsPage{})
	app.Route("/settings", &settingsPage{})
	app.Route("/stats", &statsPage{})
	app.Route("/users", &usersPage{})

	app.RunWhenOnBrowser()
}

func initServer() {}
