package frontend

import (
	"go.vxn.dev/littr/pkg/frontend/polls"
	"go.vxn.dev/littr/pkg/frontend/settings"
	"go.vxn.dev/littr/pkg/frontend/users"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

/*
 *  polls view
 */

type PollsView struct {
	app.Compo
}

func (v *PollsView) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("polls / littr")
}

func (v *PollsView) Render() app.UI {
	return app.Div().Body(
		&header{},
		&footer{},
		&polls.Content{},
	)
}

/*
 *  settings
 */

type SettingsView struct {
	app.Compo

	mode string
}

func (v *SettingsView) Render() app.UI {
	return app.Div().Body(
		&header{},
		&footer{},
		&settings.Content{},
	)
}

func (v *SettingsView) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("settings / littr")

	ctx.LocalStorage().Get("mode", &v.mode)
}

/*
 *  users view
 */

type UsersView struct {
	app.Compo
}

func (v *UsersView) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("users / littr")
}

func (v *UsersView) Render() app.UI {
	return app.Div().Body(
		&header{},
		&footer{},
		&users.Content{},
	)
}
