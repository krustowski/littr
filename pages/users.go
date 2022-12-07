package pages

import (
	"encoding/json"
	"io/ioutil"
	"litter-go/backend"
	"log"
	"net/http"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type UsersPage struct {
	app.Compo
}

type usersContent struct {
	app.Compo
	users []backend.User `json:"users"`

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

func usersAPI() *[]byte {
	// push requests use PUT method
	req, err := http.NewRequest("GET", "/api/users", nil)
	if err != nil {
		log.Print(err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}

	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		log.Print(err)
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Print(err)
	}

	return &data
}

func (c *usersContent) OnNav(ctx app.Context) {
	ctx.Async(func() {
		var usersPre backend.Users

		if uu := usersAPI(); uu != nil {

			err := json.Unmarshal(*uu, &usersPre)
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
		if ok := backend.UserFlowToggle(flowName); !ok {
			c.toastText = "generic backend error"
			return
		}

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

					return app.Tr().Body(
						app.Td().Body(
							app.B().Text(user.Nickname).Class("deep-orange-text"),
							app.Div().Class("space"),
							app.Text(user.About),
						),
						app.Td().Body(
							app.Button().Class("responsive deep-orange7 white-text bold").
								Name(user.Nickname).OnClick(c.onClick).Body(
								app.Text("flow"),
							),
						),
					)
				}),
			),
		),

		app.Div().Class("small-space"),
		app.A().Class("loader center large deep-orange"+loaderActiveClass),
	)
}
