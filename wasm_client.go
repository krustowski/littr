//go:build wasm
// +build wasm

package main

import "github.com/maxence-charriere/go-app/v9/pkg/app"

func initWASM() {
	app.Route("/flow", &pages.Flow{})
	app.Route("/login", &pages.Login{})
	app.Route("/logout", &pages.Logout{})
	app.Route("/polls", &pages.Polls{})
	app.Route("/settings", &pages.Settings{})
	app.Route("/stats", &pages.Stats{})
	app.Route("/users", &pages.Users{})

	app.RunWhenOnBrowser()
}

func initServer() {}
