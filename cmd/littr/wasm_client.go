//go:build wasm
// +build wasm

package main

import (
	fe "go.vxn.dev/littr/pkg/frontend"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func initClient() {
	app.Route("/", &fe.WelcomePage{})
	app.Route("/flow", &fe.FlowPage{})
	app.RouteWithRegexp("/flow/post/\\d+", &fe.FlowPage{})
	app.RouteWithRegexp("/flow/hashtag/\\w+", &fe.FlowPage{})
	app.RouteWithRegexp("/flow/user/\\w+", &fe.FlowPage{})
	app.Route("/login", &fe.LoginPage{})
	app.Route("/logout", &fe.LoginPage{})
	app.Route("/polls", &fe.PollsPage{})
	app.Route("/post", &fe.PostPage{})
	app.Route("/register", &fe.RegisterPage{})
	app.Route("/reset", &fe.ResetPage{})
	app.RouteWithRegexp("/reset/\\w+", &fe.ResetPage{})
	app.Route("/settings", &fe.SettingsPage{})
	app.Route("/stats", &fe.StatsPage{})
	app.Route("/tos", &fe.ToSPage{})
	app.Route("/users", &fe.UsersPage{})

	app.RunWhenOnBrowser()
}

// function initServer() is blanked here to reduce the final WASM binary file size, which is used on the client's side (see build at the top of this file)
func initServer() {}
