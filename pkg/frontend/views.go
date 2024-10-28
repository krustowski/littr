// The umbrella package for the (mainly) WASM app's client service.
package frontend

import (
	"go.vxn.dev/littr/pkg/frontend/flow"
	"go.vxn.dev/littr/pkg/frontend/login"
	"go.vxn.dev/littr/pkg/frontend/navbars"
	"go.vxn.dev/littr/pkg/frontend/polls"
	"go.vxn.dev/littr/pkg/frontend/post"
	"go.vxn.dev/littr/pkg/frontend/register"
	"go.vxn.dev/littr/pkg/frontend/reset"
	"go.vxn.dev/littr/pkg/frontend/settings"
	"go.vxn.dev/littr/pkg/frontend/stats"
	"go.vxn.dev/littr/pkg/frontend/tos"
	"go.vxn.dev/littr/pkg/frontend/users"
	"go.vxn.dev/littr/pkg/frontend/welcome"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

/*func render[T any](component interface{}) app.UI {
	compo, ok := component.(T)
	if !ok {
		return nil
	}

	return app.Div().Body(
		&navbars.Header{},
		&navbars.Footer{},
		&compo,
	)
}*/

//
//  flow view
//

type FlowView struct {
	app.Compo
}

func (v *FlowView) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("flow / littr")
}

func (v *FlowView) Render() app.UI {
	return app.Div().Body(
		&navbars.Header{},
		&navbars.Footer{},
		&flow.Content{},
	)
}

//
//  login view
//

type LoginView struct {
	app.Compo
	userLogged bool
}

func (v *LoginView) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("login / littr")
}

func (v *LoginView) Render() app.UI {
	return app.Div().Body(
		&navbars.Header{},
		&navbars.Footer{},
		&login.Content{},
	)
}

func (v *LoginView) OnMount(ctx app.Context) {
	if ctx.Page().URL().Path == "/logout" {
		// destroy auth manually without API
		//ctx.LocalStorage().Set("userLogged", false)
		//ctx.LocalStorage().Set("userName", "")
		//ctx.LocalStorage().Set("flowRecords", nil)
		ctx.SetState("user", "")
		ctx.SetState("authGranted", false)

		v.userLogged = false

		ctx.Navigate("/login")
	}
}

//
//  polls view
//

type PollsView struct {
	app.Compo
}

func (v *PollsView) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("polls / littr")
}

func (v *PollsView) Render() app.UI {
	return app.Div().Body(
		&navbars.Header{},
		&navbars.Footer{},
		&polls.Content{},
	)
}

//
//  posts view
//

type PostView struct {
	app.Compo
}

func (v *PostView) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("post / littr")
}

func (v *PostView) Render() app.UI {
	return app.Div().Body(
		&navbars.Header{},
		&navbars.Footer{},
		&post.Content{},
	)
}

//
//  register view
//

type RegisterView struct {
	app.Compo
}

func (v *RegisterView) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("register / littr")
}

func (v *RegisterView) Render() app.UI {
	return app.Div().Body(
		&navbars.Header{},
		&navbars.Footer{},
		&register.Content{},
	)
}

//
//  reset view
//

type ResetView struct {
	app.Compo
	userLogged bool
}

func (v *ResetView) OnMount(ctx app.Context) {
}

func (v *ResetView) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("reset / littr")
}

func (v *ResetView) Render() app.UI {
	return app.Div().Body(
		&navbars.Header{},
		&navbars.Footer{},
		&reset.Content{},
	)
}

//
//  settings view
//

type SettingsView struct {
	app.Compo

	mode string
}

func (v *SettingsView) Render() app.UI {
	return app.Div().Body(
		&navbars.Header{},
		&navbars.Footer{},
		&settings.Content{},
	)
}

func (v *SettingsView) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("settings / littr")

	ctx.LocalStorage().Get("mode", &v.mode)
}

//
//  stats view
//

type StatsView struct {
	app.Compo
}

func (v *StatsView) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("stats / littr")
}

func (v *StatsView) Render() app.UI {
	return app.Div().Body(
		&navbars.Header{},
		&navbars.Footer{},
		&stats.Content{},
	)
}

//
//  ToS (terms of service) view
//

type ToSView struct {
	app.Compo
}

func (v *ToSView) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("ToS / littr")
}

func (v *ToSView) Render() app.UI {
	return app.Div().Body(
		&navbars.Header{},
		&navbars.Footer{},
		&tos.Content{},
	)
}

//
//  users view
//

type UsersView struct {
	app.Compo
}

func (v *UsersView) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("users / littr")
}

func (v *UsersView) Render() app.UI {
	return app.Div().Body(
		&navbars.Header{},
		&navbars.Footer{},
		&users.Content{},
	)
}

//
//  welcome view
//

type WelcomeView struct {
	app.Compo

	mode string
}

func (v *WelcomeView) Render() app.UI {
	return app.Div().Body(
		&navbars.Header{},
		&navbars.Footer{},
		&welcome.Content{},
	)
}

func (v *WelcomeView) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("welcome / littr")
}
