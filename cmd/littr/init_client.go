package main

import (
	fe "go.vxn.dev/littr/pkg/frontend"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

// initClientCommon is a web application router initialization helper function. It maps various routes to their frontend view conterparts.
func initClientCommon() {
	app.Route("/", &fe.WelcomeView{})
	app.RouteWithRegexp("/activation/\\w+", &fe.LoginView{})
	app.Route("/flow", &fe.FlowView{})
	app.RouteWithRegexp("/flow/posts/\\d+", &fe.FlowView{})
	app.RouteWithRegexp("/flow/hashtags/\\w+", &fe.FlowView{})
	app.RouteWithRegexp("/flow/users/\\w+", &fe.FlowView{})
	app.Route("/login", &fe.LoginView{})
	app.Route("/logout", &fe.LoginView{})
	app.Route("/polls", &fe.PollsView{})
	app.RouteWithRegexp("/polls/\\d+", &fe.PollsView{})
	app.RouteWithRegexp("/polls/\\w+", &fe.PollsView{})
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
