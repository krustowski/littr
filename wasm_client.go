//go:build wasm
// +build wasm

package main

import (
	"litter-go/pages"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func initWASM() {
	app.Route("/", &pages.LoginPage{})
	app.Route("/flow", &pages.FlowPage{})
	app.Route("/login", &pages.LoginPage{})
	app.Route("/logout", &pages.LoginPage{})
	app.Route("/polls", &pages.PollsPage{})
	app.Route("/post", &pages.PostPage{})
	app.Route("/settings", &pages.SettingsPage{})
	app.Route("/stats", &pages.StatsPage{})
	app.Route("/users", &pages.UsersPage{})

	app.RunWhenOnBrowser()
}

func initServer() {}
