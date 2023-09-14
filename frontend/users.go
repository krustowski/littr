package frontend

import (
	"encoding/json"
	"log"

	"go.savla.dev/littr/config"
	"go.savla.dev/littr/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type UsersPage struct {
	app.Compo
}

type usersContent struct {
	app.Compo

	users map[string]models.User `json:"users"`

	user models.User

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
	// show loader
	c.loaderShow = true

	var enUser string
	var user models.User
	user.FlowList = make(map[string]bool)

	ctx.LocalStorage().Get("user", &enUser)

	// decode, decrypt and unmarshal the local storage string
	if err := prepare(enUser, &user); err != nil {
		c.toastText = "frontend decoding/decryption failed: " + err.Error()
		c.toastShow = true
		return
	}

	c.user = user

	//name := c.user.Nickname
	//flowList := c.user.FlowList

	ctx.Async(func() {
		usersPre := struct {
			Users map[string]models.User `json:"users"`
		}{}

		if data, ok := litterAPI("GET", "/api/users", nil); ok {
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

		// do not save new flow user to local var until it is saved on backend
		//flowRecords := append(c.flowRecords, flowName)

		if _, found := c.user.FlowList[flowName]; found {
			delete(c.user.FlowList, flowName)
		} else {
			c.user.FlowList[flowName] = true
		}

		updateData := &models.User{
			Nickname:   c.user.Nickname,
			Passphrase: c.user.Passphrase,
			About:      c.user.About,
			Email:      c.user.Email,
			FlowList:   c.user.FlowList,
		}

		respRaw, ok := litterAPI("PUT", "/api/users", updateData)
		if !ok {
			c.toastShow = true
			c.toastText = "generic backend error"
			return
		}

		response := struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		}{}
		if err := json.Unmarshal(*respRaw, &response); err != nil {
			log.Println(err.Error())
			return
		}

		if response.Code != 200 {
			c.toastShow = true
			c.toastText = "user update failed"
			return
		}

		var stream []byte
		if err := reload(c.user, &stream); err != nil {
			c.toastShow = true
			c.toastText = "local storage reload failed: " + err.Error()
			return
		}

		ctx.LocalStorage().Set("user", config.Encrypt(config.Pepper, string(stream)))

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
		app.H5().Text("littr flowers").Style("padding-top", config.HeaderTopPadding),
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
				app.Range(c.users).Map(func(key string) app.UI {
					user := c.users[key]

					var inFlow bool = false
					for key, val := range c.user.FlowList {
						if user.Nickname == key {
							inFlow = val
							break
						}
					}

					log.Println(c.user.FlowList)

					return app.Tr().Body(
						app.Td().Body(
							app.B().Text(user.Nickname).Class("deep-orange-text"),
							app.Div().Class("space"),
							app.Text(user.About),
						),
						// make button inactive for logged user
						app.If(user.Nickname == c.user.Nickname,
							app.Td().Body(
								app.Button().Class("responsive deep-orange7 white-text bold").
									Disabled(true).Body(
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
			app.Div().Class("loader center large deep-orange"+loaderActiveClass),
		),
	)
}
