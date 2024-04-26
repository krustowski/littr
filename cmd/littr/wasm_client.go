//go:build wasm
// +build wasm

package main

import (
	fe "go.savla.dev/littr/pkg/frontend"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func initClient() {
	app.Route("/", &fe.LoginPage{})
	app.Route("/flow", &fe.FlowPage{})
	app.RouteWithRegexp("/flow/post/\\d+", &fe.FlowPage{})
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

func initServer() {}
