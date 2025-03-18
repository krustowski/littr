package organisms

import (
	"time"

	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
	"go.vxn.dev/littr/pkg/models"
)

type ModalUserInfo struct {
	app.Compo

	User      models.User
	Users     map[string]models.User
	ShowModal bool

	OnClickDismissActionName  string
	OnClickUserFlowActionName string

	userRegisteredTime string
	userLastActiveTime string
}

func (m *ModalUserInfo) processTimestamps() {
	if m.User.Nickname != "" {
		registeredTime := m.User.RegisteredTime
		lastActiveTime := m.User.LastActiveTime

		registered := app.Window().
			Get("Date").
			New(registeredTime.Format(time.RFC3339))

		lastActive := app.Window().
			Get("Date").
			New(lastActiveTime.Format(time.RFC3339))

		m.userRegisteredTime = registered.Call("toLocaleString", "en-GB").String()
		m.userLastActiveTime = lastActive.Call("toLocaleString", "en-GB").String()
	}
}

func (m *ModalUserInfo) Render() app.UI {
	m.processTimestamps()

	return app.Div().Body(
		app.If(m.ShowModal, func() app.UI {
			return app.Dialog().ID("user-modal").Class("grey10 white-text center-align active thicc").Style("max-width", "90%").Body(

				&atoms.Image{
					Class:  "",
					Src:    m.User.AvatarURL,
					Styles: map[string]string{"max-width": "120px", "border-radius": "50%"},
				},

				app.Div().Class("row center-align").Body(
					app.H5().Class().Body(
						app.A().Href("/flow/users/"+m.User.Nickname).Text(m.User.Nickname),
					),

					app.If(m.User.Web != "", func() app.UI {
						return app.A().Href(m.User.Web).Body(
							app.Span().Class("bold").Body(
								app.I().Text("captive_portal"),
							),
						)
					}),
				),

				app.If(m.User.About != "", func() app.UI {
					return app.Article().Class("center-align white-text border thicc").Style("word-break", "break-word").Style("hyphens", "auto").Text(m.User.About)
				}),

				app.Article().Class("white-text border left-align thicc").Body(
					app.P().Class("bold").Text("Registered"),
					app.P().Class().Text(m.userRegisteredTime),

					app.P().Class("bold").Text("Last online"),
					app.P().Class().Text(m.userLastActiveTime),
				),

				//app.Div().Class("large-space"),
				app.Div().Class("row center-align").Body(
					&atoms.Button{
						Class:             "max black white-text thicc",
						Icon:              "close",
						Text:              "Close",
						OnClickActionName: m.OnClickDismissActionName,
					},

					&atoms.Button{
						ID:                m.User.Nickname,
						Class:             "max primary-container white-text thicc",
						Icon:              "tsunami",
						Text:              "Flow",
						OnClickActionName: m.OnClickUserFlowActionName,
					},
				),
			)
		}),
	)
}
