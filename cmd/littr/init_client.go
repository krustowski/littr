//go:build wasm || server
// +build wasm server

package main

import (
	fe "go.vxn.dev/littr/pkg/frontend"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

// initClientCommon is a web application router initialization helper function. It maps various routes to their frontend view conterparts.
func initClientCommon() {
	app.Route("/", func() app.Composer {
		return &fe.WelcomeView{}
	})
	app.RouteWithRegexp("^/activation/[a-zA-Z]+$", func() app.Composer {
		return &fe.LoginView{}
	})
	app.Route("/flow", func() app.Composer {
		return &fe.FlowView{}
	})
	app.RouteWithRegexp("^/flow/posts/[0-9]+$", func() app.Composer {
		return &fe.FlowView{}
	})
	app.RouteWithRegexp("^/flow/hashtags/[a-zA-Z]+$", func() app.Composer {
		return &fe.FlowView{}
	})
	app.RouteWithRegexp("^/flow/users/[a-zA-Z]+$", func() app.Composer {
		return &fe.FlowView{}
	})
	app.Route("/login", func() app.Composer {
		return &fe.LoginView{}
	})
	app.Route("/logout", func() app.Composer {
		return &fe.LoginView{}
	})
	app.Route("/polls", func() app.Composer {
		return &fe.PollsView{}
	})
	app.RouteWithRegexp("^/polls/[0-9]+$", func() app.Composer {
		return &fe.PollsView{}
	})
	app.RouteWithRegexp("^/polls/[a-zA-Z]+$", func() app.Composer {
		return &fe.PollsView{}
	})
	app.Route("/post", func() app.Composer {
		return &fe.PostView{}
	})
	app.Route("/register", func() app.Composer {
		return &fe.RegisterView{}
	})
	app.Route("/reset", func() app.Composer {
		return &fe.ResetView{}
	})
	app.RouteWithRegexp("^/reset/[a-zA-Z]+$", func() app.Composer {
		return &fe.ResetView{}
	})
	app.Route("/settings", func() app.Composer {
		return &fe.SettingsView{}
	})
	app.Route("/stats", func() app.Composer {
		return &fe.StatsView{}
	})
	app.RouteWithRegexp("^/success/[a-zA-Z]+$", func() app.Composer {
		return &fe.LoginView{}
	})
	app.Route("/tos", func() app.Composer {
		return &fe.ToSView{}
	})
	app.Route("/users", func() app.Composer {
		return &fe.UsersView{}
	})

	app.RunWhenOnBrowser()
}
