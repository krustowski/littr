//go:build wasm
// +build wasm

package main

import (
	fe "go.vxn.dev/littr/pkg/frontend"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func initClient() {
	app.Route("/", &fe.WelcomeView{})
	app.Route("/flow", &fe.FlowPage{})
	app.RouteWithRegexp("/flow/post/\\d+", &fe.FlowPage{})
	app.RouteWithRegexp("/flow/hashtag/\\w+", &fe.FlowPage{})
	app.RouteWithRegexp("/flow/user/\\w+", &fe.FlowPage{})
	app.Route("/login", &fe.LoginPage{})
	app.Route("/logout", &fe.LoginPage{})
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

// function initServer() is blanked here to reduce the final WASM binary file size, which is used on the client's side (see build at the top of this file)
func initServer() {}
