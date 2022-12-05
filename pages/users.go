package pages

import (
	"litter-go/backend"
	"log"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type UsersPage struct {
	app.Compo
}

type usersContent struct {
	app.Compo
	users []backend.User
}

func (p *UsersPage) OnNav(ctx app.Context) {
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

func (c *usersContent) OnNav(ctx app.Context) {
	ctx.Async(func() {
		var users []backend.User
		if uu := backend.GetUsers(); uu != nil {
			users = *uu

			// Storing HTTP response in component field:
			ctx.Dispatch(func(ctx app.Context) {
				c.users = users
				log.Println("dispatch ends")
			})
		}
	})
}

func (c *usersContent) Render() app.UI {
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
							app.B().Text(user.Nickname).Class("deep-orange-text"),
							app.Div().Class("space"),
							app.Text(user.About),
						),
						app.Td().Body(
							app.Button().Class("responsive deep-orange7 white-text bold").Body(
								app.Text("flow"),
							),
						),
					)
				}),
			),
		),
	)
}
