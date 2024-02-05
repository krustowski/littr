//go:build wasm
// +build wasm

package main

import (
	frontend "go.savla.dev/littr/frontend"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func initClient() {
	app.Route("/", &frontend.LoginPage{})
	app.Route("/flow", &frontend.FlowPage{})
	app.RouteWithRegexp("/flow/post/\\d+", &frontend.FlowPage{})
	app.RouteWithRegexp("/flow/user/\\w+", &frontend.FlowPage{})
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

func initServer() {}
