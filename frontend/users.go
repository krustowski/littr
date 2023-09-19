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

	//user.FlowList = make(map[string]bool)
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

func (c *usersContent) OnMount(ctx app.Context) {
	ctx.Handle("toggle", c.handleToggle)
}

func (c *usersContent) handleToggle(ctx app.Context, a app.Action) {
	key, ok := a.Value.(string)
	if !ok {
		return
	}

	flowList := c.user.FlowList

	if flowList == nil {
		flowList = make(map[string]bool)
		flowList[c.user.Nickname] = true
		c.user.FlowList = flowList
	}

	if _, found := flowList[key]; found {
		flowList[key] = !flowList[key]
	} else {
		flowList[key] = true
	}

	c.user.FlowList = flowList

	ctx.Async(func() {
		// do not save new flow user to local var until it is saved on backend
		//flowRecords := append(c.flowRecords, flowName)

		updateData := &models.User{
			Nickname:   c.user.Nickname,
			Passphrase: c.user.Passphrase,
			About:      c.user.About,
			Email:      c.user.Email,
			FlowList:   flowList,
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
			c.toastShow = true
			c.toastText = "user update failed: " + err.Error()
			return
		}

		if response.Code != 200 && response.Code != 201 {
			c.toastShow = true
			c.toastText = "user update failed: " + response.Message
			log.Println(response.Message)
			return
		}

		var stream []byte
		if err := reload(c.user, &stream); err != nil {
			c.toastShow = true
			c.toastText = "local storage reload failed: " + err.Error()
			return
		}

		ctx.LocalStorage().Set("user", stream)

		c.toastShow = false
	})
}

func (c *usersContent) onClick(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()
	ctx.NewActionWithValue("toggle", key)
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

		app.Table().Class("border left-align").Style("max-width", "100%").Body(
			app.THead().Body(
				app.Tr().Body(
					app.Th().Text("nick, about"),
					app.Th().Text("flow list"),
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

					return app.Tr().Body(
						app.Td().Style("max-width", "0").Style("text-overflow", "-").Style("overflow", "clip").Style("word-break", "break-all").Style("hyphens", "auto").Body(
							app.P().Text(user.Nickname).Class("deep-orange-text bold"),
							app.Div().Class("space"),
							app.P().Style("word-break", "break-word").Style("hyphens", "auto").Text(user.About),
						),

						// make button inactive for logged user
						app.If(user.Nickname == c.user.Nickname,
							app.Td().Body(
								app.Button().Class("responsive deep-orange7 white-text bold").Disabled(true).Body(
									app.Text("that's you"),
								),
							),

						// toggle off
						).ElseIf(inFlow,
							app.Td().Style("max-width", "50%").Body(
								//app.Button().Class("responsive black white-border white-text bold left-shadow").ID(user.Nickname).OnClick(c.onClick).Body(
								app.Button().Class("responsive black white-border white-text bold left-shadow").ID(user.Nickname).OnClick(c.onClick).Body(
									app.I().Text("close"),
									app.Text("toggle off"),
								),
							),

						// toggle on
						).Else(
							app.Td().Style("max-width", "50%").Body(
								app.Button().Class("responsive deep-orange7 white-text bold").ID(user.Nickname).OnClick(c.onClick).Body(
									app.I().Text("done"),
									app.Text("toggle on"),
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
