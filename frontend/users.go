package frontend

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

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

	user        models.User
	userInModal models.User

	loaderShow bool

	toastShow bool
	toastText string

	usersButtonDisabled  bool
	showUserPreviewModal bool
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

		// delet dis
		for k, u := range usersPre.Users {
			u.Searched = true
			usersPre.Users[k] = u
		}

		// Storing HTTP response in component field:
		ctx.Dispatch(func(ctx app.Context) {
			c.user = user
			c.users = usersPre.Users

			c.loaderShow = false
		})
	})
}

func (c *usersContent) OnMount(ctx app.Context) {
	ctx.Handle("toggle", c.handleToggle)
	ctx.Handle("search", c.handleSearch)
	ctx.Handle("preview", c.handleUserPreview)
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
		//c.user.FlowList = flowList
	}

	if _, found := flowList[key]; found {
		flowList[key] = !flowList[key]
	} else {
		flowList[key] = true
	}

	ctx.Async(func() {
		toastText := ""

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
			toastText = "generic backend error"
			return
		}

		response := struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		}{}

		if err := json.Unmarshal(*respRaw, &response); err != nil {
			toastText = "user update failed: " + err.Error()
			return
		}

		if response.Code != 200 && response.Code != 201 {
			toastText = "user update failed: " + response.Message
			log.Println(response.Message)
			return
		}

		var stream []byte
		if err := reload(c.user, &stream); err != nil {
			toastText = "local storage reload failed: " + err.Error()
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			ctx.LocalStorage().Set("user", stream)

			c.toastText = toastText
			c.toastShow = (toastText != "")
			c.usersButtonDisabled = false

			c.user.FlowList = flowList
			log.Println("dispatch ends")
		})
	})
}

func (c *usersContent) onSearch(ctx app.Context, e app.Event) {
	val := ctx.JSSrc().Get("value").String()

	if len(val) > 20 {
		return
	}

	ctx.NewActionWithValue("search", val)
}

func (c *usersContent) handleSearch(ctx app.Context, a app.Action) {
	val, ok := a.Value.(string)
	if !ok {
		return
	}

	ctx.Async(func() {
		users := c.users

		// iterate over calculated stats' "rows" and find matchings
		for key, user := range users {
			//user := users[key]
			user.Searched = false

			if strings.Contains(key, val) {
				log.Println(key)
				user.Searched = true
			}

			users[key] = user
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.users = users

			c.loaderShow = false
		})
		return
	})
}

func (c *usersContent) handleUserPreview(ctx app.Context, a app.Action) {
	val, ok := a.Value.(string)
	if !ok {
		return
	}

	ctx.Async(func() {
		user := c.users[val]

		ctx.Dispatch(func(ctx app.Context) {
			c.showUserPreviewModal = true
			c.userInModal = user
		})
	})
	return
}

func (c *usersContent) onClickUser(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()
	ctx.NewActionWithValue("preview", key)
	c.usersButtonDisabled = true
	c.showUserPreviewModal = true
}

func (c *usersContent) onClick(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()
	ctx.NewActionWithValue("toggle", key)
	c.usersButtonDisabled = true
}

func (c *usersContent) dismissToast(ctx app.Context, e app.Event) {
	c.toastText = ""
	c.toastShow = (c.toastText != "")
	c.usersButtonDisabled = false
	c.showUserPreviewModal = false
}

func (c *usersContent) getGravatarURL() string {
	// TODO: do not hardcode this
	baseURL := "https://littr.n0p.cz/"
	email := strings.ToLower(c.userInModal.Email)
	size := 150

	defaultImage := "/web/android-chrome-192x192.png"

	byteEmail := []byte(email)
	hashEmail := md5.Sum(byteEmail)
	hashedStringEmail := hex.EncodeToString(hashEmail[:])

	url := "https://www.gravatar.com/avatar/" + hashedStringEmail + "?d=" + baseURL + "&s=" + strconv.Itoa(size)

	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		return defaultImage
	}

	return url
}

func (c *usersContent) Render() app.UI {
	users := c.users
	userGravatarURL := ""

	var userInModalInfo map[string]string = nil

	if c.showUserPreviewModal {
		userInModalInfo = map[string]string{
			"full name": c.userInModal.FullName,
			"web":       c.userInModal.Web,
			//"e-mail":    c.userInModal.Email,
			"last active": c.userInModal.LastActiveTime.String(),
			"registered":  c.userInModal.RegisteredTime.String(),
		}

		//userGravatarURL := getGravatar(c.userInModal.Email)
		userGravatarURL = c.getGravatarURL()
	}

	return app.Main().Class("responsive").Body(
		app.H5().Text("littr flowers").Style("padding-top", config.HeaderTopPadding),
		app.P().Text("simplified user table, available to add to the flow!"),
		app.Div().Class("space"),

		// snackbar
		app.A().OnClick(c.dismissToast).Body(
			app.If(c.toastText != "",
				app.Div().Class("snackbar red10 white-text top active").Body(
					app.I().Text("error"),
					app.Span().Text(c.toastText),
				),
			),
		),

		// user info modal
		app.If(c.showUserPreviewModal && userGravatarURL != "" && userInModalInfo != nil,
			app.Dialog().Class("grey9 white-text center-align active").Style("max-width", "90%").Body(

				app.Img().Class("small-width small-heigh").Src(userGravatarURL),

				app.Nav().Class("center-align").Body(
					app.H5().Text(c.userInModal.Nickname),
				),

				app.Div().Class("space"),

				app.If(c.userInModal.About != "",
					app.Article().Class("center-align").Style("word-break", "break-word").Style("hyphens", "auto").Text(c.userInModal.About),
					app.Div().Class("space"),
				),

				app.Table().Class("border center-align").Style("max-width", "100%").Body(
					app.TBody().Body(
						app.Range(userInModalInfo).Map(func(key string) app.UI {
							if userInModalInfo[key] == "" {
								return app.Text("")
							}

							return app.Tr().Body(
								app.Td().Body(
									app.Text(key),
								),

								app.If(key == "web",
									app.Td().Body(
										app.A().Class("bold").Href(userInModalInfo[key]).Text(userInModalInfo[key]),
									),
								).Else(
									app.Td().Body(
										app.Span().Class("bold").Text(userInModalInfo[key]),
									),
								),
							)
						}),
					),
				),

				//app.Div().Class("large-space"),
				app.Nav().Class("center-align").Body(
					app.Button().Class("border deep-orange7 white-text").Text("close").OnClick(c.dismissToast),
				),
			),
		),

		app.Div().Class("field prefix round fill").Body(
			app.I().Class("front").Text("search"),
			//app.Input().Type("search").OnChange(c.ValueTo(&c.searchString)).OnSearch(c.onSearch),
			app.Input().Type("text").OnChange(c.onSearch).OnSearch(c.onSearch),
		),

		app.Table().Class("border left-align").Style("max-width", "100%").Body(
			app.THead().Body(
				app.Tr().Body(
					app.Th().Text("nick, about"),
					app.Th().Text("flow list"),
				),
			),
			app.TBody().Body(
				app.Range(users).Map(func(key string) app.UI {
					user := users[key]

					var inFlow bool = false

					for key, val := range c.user.FlowList {
						if user.Nickname == key {
							inFlow = val
							break
						}
					}

					if !user.Searched || user.Nickname == "system" {
						return app.Text("")
					}

					return app.Tr().Body(
						app.Td().Style("max-width", "0").Style("text-overflow", "-").Style("overflow", "clip").Style("word-break", "break-all").Style("hyphens", "auto").Body(
							app.P().ID(user.Nickname).Text(user.Nickname).Class("deep-orange-text bold").OnClick(c.onClickUser),
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
						).ElseIf(user.Nickname == "system",
							app.Td().Body(
								app.Button().Class("responsive deep-orange7 white-text bold").Disabled(true).Body(
									app.Text("system acc"),
								),
							),

						// toggle off
						).ElseIf(inFlow,
							app.Td().Style("max-width", "50%").Body(
								//app.Button().Class("responsive black white-border white-text bold left-shadow").ID(user.Nickname).OnClick(c.onClick).Body(
								app.Button().Class("border responsive black white-border white-text bold left-shadow").ID(user.Nickname).OnClick(c.onClick).Disabled(c.usersButtonDisabled).Body(
									//app.I().Text("done"),
									app.Text("remove from flow"),
								),
							),

						// toggle on
						).Else(
							app.Td().Style("max-width", "50%").Body(
								app.Button().Class("responsive deep-orange7 white-text bold").ID(user.Nickname).OnClick(c.onClick).Disabled(c.usersButtonDisabled).Body(
									//app.I().Text("done"),
									app.Text("add to flow"),
								),
							),
						),
					)
				}),
			),
		),

		app.If(c.loaderShow,
			app.Div().Class("small-space"),
			app.Div().Class("loader center large deep-orange active"),
		),
	)
}
