package pages

import "github.com/maxence-charriere/go-app/v9/pkg/app"

type UsersPage struct {
	app.Compo
}

type usersContent struct {
	app.Compo
	users []User
}

type User struct {
	Nickname string
	About    string
	Flow     []string
}

func (p *SettingsPage) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("users / littr")
}

func (p *UsersPage) Render() app.UI {
	return app.Div().Body(
		app.Body().Class("dark"),
		&header{},
		&usersContent{},
		&footer{},
	)
}

func (c *usersContent) Render() app.UI {
	c.users = []User{
		{Nickname: "krusty", About: "idk lemme just die ffs frfr"},
		{Nickname: "lmao", About: "wtf is this site lmao"},
	}

	return app.Main().Class("responsive").Body(
		app.H5().Text("littr user list"),
		app.P().Text("simplified user table, available to add to the flow!"),
		app.Div().Class("space"),

		app.Table().Class("border left-align").Body(
			app.THead().Body(
				app.Tr().Body(
					app.Th().Text("nick, about"),
					app.Th().Text("flow"),
				),
			),
			app.TBody().Body(
				app.Range(c.users).Slice(func(i int) app.UI {
					user := c.users[i]

					return app.Tr().Body(
						app.Td().Body(
							app.B().Text(user.Nickname).Style("color", "orange"),
							app.Div().Class("space"),
							app.Text(user.About),
						),
						app.Td().Body(
							app.Button().Class().Body(
								app.Text("flow"),
							),
						),
					)
				}),
			),
		),
	)
}
