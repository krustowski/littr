// The settings view and view-controllers logic package.
package settings

import (
	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/helpers"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Content struct {
	app.Compo

	// TODO: review this
	loggedUser string

	// used with forms
	passphrase        string
	passphraseAgain   string
	passphraseCurrent string
	aboutText         string
	website           string

	// loaded logged user's struct
	user models.User

	// message toast fields
	toast common.Toast

	darkModeOn   bool
	replyNotifOn bool

	notificationPermission app.NotificationPermission
	subscribed             bool
	subscription           struct {
		Replies  bool
		Mentions bool
	}

	settingsButtonDisabled bool

	deleteAccountModalShow      bool
	deleteSubscriptionModalShow bool

	VAPIDpublic string
	devices     []models.Device
	thisDevice  models.Device

	thisDeviceUUID string
	interactedUUID string

	themeMode string

	//keyDownEventListener func()
}

func (c *Content) OnMount(ctx app.Context) {
	if app.IsServer {
		return
	}

	// This function call is broken due to the slider not hitting the actual top of the page.
	//ctx.ScrollTo("anchor-settings-top")
	//scrollObj := map[string]any{"top": 0}
	//app.Window().Call("scrollTo", app.ValueOf(scrollObj))

	c.notificationPermission = ctx.Notifications().Permission()

	ctx.Handle("dismiss", c.handleDismiss)

	ctx.Handle("options-switch-change", c.handleOptionsSwitchChange)
	ctx.Handle("notifs-switch-change", c.handleNotificationSwitchChange)

	ctx.Handle("subscription-delete-modal-show", c.handleModalShow)
	ctx.Handle("user-delete-modal-show", c.handleModalShow)

	ctx.Handle("avatar-change", c.handleImageUpload)
	ctx.Handle("user-delete", c.handleUserDelete)
	ctx.Handle("subscription-delete", c.handleSubscriptionDelete)

	ctx.Handle("passphrase-submit", c.handlePassphraseChange)
	ctx.Handle("about-you-submit", c.handleOptionsChange)
	ctx.Handle("website-submit", c.handleOptionsChange)
}

func (c *Content) OnNav(ctx app.Context) {
	if app.IsServer {
		return
	}

	toast := common.Toast{AppContext: &ctx}

	ctx.Dispatch(func(ctx app.Context) {
		c.settingsButtonDisabled = true
	})

	ctx.ScrollTo("anchor-settings-top")

	ctx.Async(func() {
		input := &common.CallInput{
			Method: "GET",
			Url:    "/api/v1/users/caller",
			PageNo: 0,
		}

		type dataModel struct {
			PublicKey string          `json:"public_key"`
			User      models.User     `json:"user"`
			Devices   []models.Device `json:"devices"`
		}

		output := &common.Response{Data: &dataModel{}}

		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		if output.Code == 401 {
			_ = ctx.LocalStorage().Set("user", "")
			_ = ctx.LocalStorage().Set("authGranted", false)

			toast.Text(common.ERR_LOGIN_AGAIN).Type(common.TTYPE_INFO).Dispatch()
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		data, ok := output.Data.(*dataModel)
		if !ok {
			toast.Text(common.ERR_CANNOT_GET_DATA).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		var thisDevice models.Device
		for _, dev := range data.User.Devices {
			if dev.UUID == ctx.DeviceID() {
				thisDevice = dev
				break
			}
		}

		subscription := c.subscription
		if helpers.Contains(thisDevice.Tags, "reply") {
			subscription.Replies = true
		}

		if helpers.Contains(thisDevice.Tags, "mention") {
			subscription.Mentions = true
		}

		ctx.SetState(common.StateNameUser, data.User)

		ctx.Dispatch(func(ctx app.Context) {
			c.user = data.User
			c.loggedUser = data.User.Nickname
			c.devices = data.User.Devices

			//c.subscribed = output.Subscribed
			c.subscription = subscription

			c.aboutText = data.User.About
			c.website = data.User.Web

			c.VAPIDpublic = data.PublicKey
			c.thisDeviceUUID = ctx.DeviceID()
			c.thisDevice = thisDevice

			c.replyNotifOn = c.notificationPermission == app.NotificationGranted
			//c.replyNotifOn = user.ReplyNotificationOn

			c.settingsButtonDisabled = false
		})
	})
}
