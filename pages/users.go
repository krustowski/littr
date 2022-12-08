package pages

import (
	"encoding/json"
	"litter-go/backend"
	"log"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type UsersPage struct {
	app.Compo
}

type usersContent struct {
	app.Compo
	users []backend.User `json:"users"`

	loggedUser  string
	flowRecords []string

	loaderShow bool

	toastShow bool
	toastText string
}

func (p *UsersPage) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("users / littr")
}

func (p *UsersPage) Render() app.UI {
	return app.Div().Body(
		app.Body().Class("dark"),
		&header{},
		&footer{},
		&usersContent{},
	)
}

func (c *usersContent) OnNav(ctx app.Context) {
	c.loaderShow = true

	ctx.LocalStorage().Get("userName", &c.loggedUser)
	ctx.LocalStorage().Get("flowRecords", &c.flowRecords)

	ctx.Async(func() {
		var usersPre backend.Users

		if data, _ := litterAPI("GET", "/api/users", nil); data != nil {
			err := json.Unmarshal(*data, &usersPre)
			if err != nil {
				log.Println(err.Error())
				return
			}
		} else {
			log.Println("cannot fetch user list")
			return
		}

		// Storing HTTP response in component field:
		ctx.Dispatch(func(ctx app.Context) {
			c.users = usersPre.Users

			c.loaderShow = false
			log.Println("dispatch ends")
		})
	})
}

func (c *usersContent) onClick(ctx app.Context, e app.Event) {
	ctx.Async(func() {
		flowName := ctx.JSSrc().Get("name").String()
		log.Println("toggle flow user: " + flowName)

		c.toastShow = true
		return
		//if ok := backend.UserFlowToggle(flowName); !ok {
		//	c.toastText = "generic backend error"
		//	return
		//}

		c.toastShow = false
		ctx.Navigate("/flow")
	})
}

func (c *usersContent) dismissToast(ctx app.Context, e app.Event) {
	c.toastShow = false
}

func (c *usersContent) Render() app.UI {
	toastActiveClass := ""
	if c.toastShow {
		toastActiveClass = " active"
	}

	loaderActiveClass := ""
	if c.loaderShow {
		loaderActiveClass = " active"
	}

	return app.Main().Class("responsive").Body(
		app.H5().Text("littr user list"),
		app.P().Text("simplified user table, available to add to the flow!"),
		app.Div().Class("space"),

		app.A().OnClick(c.dismissToast).Body(
			app.Div().Class("toast red10 white-text top"+toastActiveClass).Body(
				app.I().Text("error"),
				app.Span().Text(c.toastText),
			),
		),

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

					var inFlow bool = false
					for _, rec := range c.flowRecords {
						log.Println(rec)
						if user.Nickname == rec {
							inFlow = true
							break
						}
					}

					log.Println(c.flowRecords)

					return app.Tr().Body(
						app.Td().Body(
							app.B().Text(user.Nickname).Class("deep-orange-text"),
							app.Div().Class("space"),
							app.Text(user.About),
						),
						// make button inactive for logged user
						app.If(user.Nickname == c.loggedUser,
							app.Td().Body(
								app.Button().Class("responsive grey white-text bold").
									Name(user.Nickname).Body(
									app.Text("that's you"),
								),
							),
						//
						).ElseIf(inFlow,
							app.Td().Body(
								app.Button().Class("responsive deep-orange7 white-text bold").
									Name(user.Nickname).OnClick(c.onClick).Body(
									app.Text("flow off"),
								),
							),
						//
						).Else(
							app.Td().Body(
								app.Button().Class("responsive deep-orange7 white-text bold").
									Name(user.Nickname).OnClick(c.onClick).Body(
									app.Text("flow on"),
								),
							),
						),
					)
				}),
			),
		),

		app.If(c.loaderShow,
			app.Div().Class("small-space"),
			app.A().Class("loader center large deep-orange"+loaderActiveClass),
		),
	)
}
