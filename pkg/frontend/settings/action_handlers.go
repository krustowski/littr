package settings

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/models"
)

// handleDismiss()
func (c *Content) handleDismiss(ctx app.Context, a app.Action) {
	ctx.Dispatch(func(ctx app.Context) {
		c.toast.TText = ""
		c.toast.TType = ""

		c.settingsButtonDisabled = false
		c.deleteAccountModalShow = false
		c.deleteSubscriptionModalShow = false
	})
}

func (c *Content) handleImageUpload(ctx app.Context, a app.Action) {
	ctx.Dispatch(func(ctx app.Context) {
		c.settingsButtonDisabled = true
	})

	callback := func() {
		ctx.Dispatch(func(ctx app.Context) {
			c.settingsButtonDisabled = false
		})
	}

	common.HandleImageUpload(ctx, a, &c.user, callback)

}

func (c *Content) handleModalShow(ctx app.Context, a app.Action) {
	id, ok := a.Value.(string)
	if !ok {
		return
	}

	switch a.Name {
	case "user-delete-modal-show":
		ctx.Dispatch(func(ctx app.Context) {
			c.deleteAccountModalShow = true
			c.settingsButtonDisabled = true
		})

	case "subscription-delete-modal-show":
		ctx.Dispatch(func(ctx app.Context) {
			c.deleteSubscriptionModalShow = true
			c.settingsButtonDisabled = true
			c.interactedUUID = id
		})
	}

}

func (c *Content) handleNotificationSwitchChange(ctx app.Context, a app.Action) {
	key, ok := a.Value.(string)
	if !ok {
		return
	}

	ctx.Dispatch(func(ctx app.Context) {
		c.settingsButtonDisabled = true
	})

	type notifSubscription struct {
		Reply   bool
		Mention bool
	}

	var tag string

	subStates := notifSubscription{
		Reply:   c.subscription.Replies,
		Mention: c.subscription.Mentions,
	}

	subscribedCurrent := func() bool {
		if subStates.Reply || subStates.Mention {
			return true
		}

		return false
	}()

	switch key {
	case "reply-notif-switch":
		subStates.Reply = !subStates.Reply
		tag = "reply"
	case "mention-notif-switch":
		subStates.Mention = !subStates.Mention
		tag = "mention"
	}

	subscribedNew := func() bool {
		if subStates.Reply || subStates.Mention {
			return true
		}

		return false
	}()

	ctx.Async(func() {
		defer ctx.Dispatch(func(ctx app.Context) {
			c.settingsButtonDisabled = false
		})

		// Unsubscribing.
		if subscribedCurrent && !subscribedNew {
			//
			c.deleteSubscription(ctx)
			return
		}

		// Subscribing.
		if !subscribedCurrent && subscribedNew {
			if !c.checkPermission(ctx) {
				return
			}

			c.createSubscription(ctx, tag)
			return
		}

		c.updateSubscriptionTag(ctx, tag)
	})
}

func (c *Content) handleOptionsChange(ctx app.Context, a app.Action) {
	ctx.Dispatch(func(ctx app.Context) {
		c.settingsButtonDisabled = true
	})

	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		defer ctx.Dispatch(func(ctx app.Context) {
			c.settingsButtonDisabled = false
		})

		var message string
		payload := c.prefillPayload()

		switch a.Name {
		case "about-you-submit":
			aboutText := strings.TrimSpace(c.aboutText)

			if aboutText == "" || aboutText == c.user.About {
				toast.Text(common.ERR_ABOUT_TEXT_UNCHANGED).Type(common.TTYPE_ERR).Dispatch()
				return
			}

			if len(aboutText) > 100 {
				toast.Text(common.ERR_ABOUT_TEXT_CHAR_LIMIT).Type(common.TTYPE_ERR).Dispatch()
				return
			}

			message = common.MSG_ABOUT_TEXT_UPDATED
			payload.AboutText = aboutText

		case "website-submit":
			websiteCompo := app.Window().GetElementByID("website-input")
			if websiteCompo.IsNull() {
				return
			}

			website := strings.TrimSpace(websiteCompo.Get("value").String())

			// check the trimmed version of website string
			if website == "" {
				toast.Text(common.ERR_WEBSITE_UNCHANGED).Type(common.TTYPE_ERR).Dispatch()
				return
			}

			// check the URL/URI format
			if _, err := url.ParseRequestURI(website); err != nil {
				toast.Text(common.ERR_WEBSITE_INVALID).Type(common.TTYPE_ERR).Dispatch()
				return
			}

			// create a regex object
			regex, err := regexp.Compile("^(http|https)://")
			if err != nil {
				toast.Text(common.ERR_WEBSITE_REGEXP_FAIL).Type(common.TTYPE_ERR).Dispatch()
				return
			}

			if !regex.MatchString(website) {
				website = "https://" + website
			}

			message = common.MSG_WEBSITE_UPDATED
			payload.WebsiteLink = website
		}

		input := &common.CallInput{
			Method:      "PATCH",
			Url:         "/api/v1/users/" + c.user.Nickname + "/options",
			Data:        payload,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		output := &common.Response{}

		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		// Dispatch the good news to client.
		ctx.Dispatch(func(ctx app.Context) {
			c.updateOptions(payload)

			// Update the LocalStorage.
			ctx.SetState(common.StateNameUser, c.user).Persist()
		})

		toast.Text(message).Type(common.TTYPE_SUCCESS).Dispatch()
	})
}

func (c *Content) handleOptionsSwitchChange(ctx app.Context, a app.Action) {
	key, ok := a.Value.(string)
	if !ok {
		return
	}

	ctx.Dispatch(func(ctx app.Context) {
		c.settingsButtonDisabled = true
	})

	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		defer ctx.Dispatch(func(ctx app.Context) {
			c.settingsButtonDisabled = false
		})

		var message string

		// See options.go.
		payload := c.prefillPayload()

		switch key {
		case "ui-mode-switch":
			message = common.MSG_UI_MODE_TOGGLE
			payload.UIMode = !payload.UIMode

		case "ui-theme-switch":
			message = common.MSG_UI_THEME_TOGGLE
			payload.UITheme = func() models.Theme {
				// Very nasty hack, but whatever.
				switch payload.UITheme {
				case models.ThemeOrang:
					return models.ThemeDefault
				default:
					return models.ThemeOrang
				}
			}()

		case "local-time-mode-switch":
			message = common.MSG_LOCAL_TIME_TOGGLE
			payload.LocalTimeMode = !payload.LocalTimeMode

		case "live-mode-switch":
			message = common.MSG_LIVE_MODE_TOGGLE
			payload.LiveMode = !payload.LiveMode

		case "private-mode-switch":
			message = common.MSG_PRIVATE_MODE_TOGGLE
			payload.Private = !payload.Private
		}

		input := &common.CallInput{
			Method:      "PATCH",
			Url:         "/api/v1/users/" + c.user.Nickname + "/options",
			Data:        payload,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		output := &common.Response{}

		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		// Dispatch the good news to client.
		ctx.Dispatch(func(ctx app.Context) {
			c.updateOptions(payload)
			c.themeMode = payload.UITheme.Bg()

			// Update the LocalStorage.
			ctx.SetState(common.StateNameUser, c.user).Persist()
		})

		toast.Text(message).Type(common.TTYPE_SUCCESS).Dispatch()
	})
}

func (c *Content) handlePassphraseChange(ctx app.Context, a app.Action) {
	ctx.Dispatch(func(ctx app.Context) {
		c.settingsButtonDisabled = true
	})

	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		defer ctx.Dispatch(func(ctx app.Context) {
			c.settingsButtonDisabled = false
		})

		passphraseCurrentCompo := app.Window().GetElementByID("passphrase-current")
		if passphraseCurrentCompo.IsNull() {
			return
		}
		passphraseNewCompo := app.Window().GetElementByID("passphrase-new")
		if passphraseNewCompo.IsNull() {
			return
		}
		passphraseAgainCompo := app.Window().GetElementByID("passphrase-new-again")
		if passphraseAgainCompo.IsNull() {
			return
		}

		passphraseCurrent := strings.TrimSpace(passphraseCurrentCompo.Get("value").String())
		passphraseNew := strings.TrimSpace(passphraseNewCompo.Get("value").String())
		passphraseAgain := strings.TrimSpace(passphraseAgainCompo.Get("value").String())

		//passphrase := strings.TrimSpace(c.passphrase)
		//passphraseAgain := strings.TrimSpace(c.passphraseAgain)
		//passphraseCurrent := strings.TrimSpace(c.passphraseCurrent)

		if passphraseNew == "" || passphraseAgain == "" || passphraseCurrent == "" {
			toast.Text(common.ERR_PASSPHRASE_MISSING).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		if passphraseNew != passphraseAgain {
			toast.Text(common.ERR_PASSPHRASE_MISMATCH).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		payload := struct {
			NewPassphrase     string `json:"new_passphrase_plain"`
			CurrentPassphrase string `json:"current_passphrase_plain"`
		}{
			NewPassphrase:     passphraseNew,
			CurrentPassphrase: passphraseCurrent,
		}

		input := &common.CallInput{
			Method:      "PATCH",
			Url:         "/api/v1/users/" + c.user.Nickname + "/passphrase",
			Data:        payload,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		output := &common.Response{}

		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		toast.Text(common.MSG_PASSPHRASE_UPDATED).Type(common.TTYPE_SUCCESS).Dispatch()
	})
}

func (c *Content) handleSubscriptionDelete(ctx app.Context, _ app.Action) {
	toast := common.Toast{AppContext: &ctx}

	uuid := c.interactedUUID
	if uuid == "" {
		toast.Text(common.ERR_SUBSCRIPTION_BLANK_UUID).Type(common.TTYPE_ERR).Dispatch()
		return
	}

	ctx.Dispatch(func(ctx app.Context) {
		c.settingsButtonDisabled = false
	})

	ctx.Async(func() {
		defer ctx.Dispatch(func(ctx app.Context) {
			c.settingsButtonDisabled = false
		})

		payload := struct {
			UUID string `json:"device_uuid"`
		}{
			UUID: uuid,
		}

		input := &common.CallInput{
			Method:      "DELETE",
			Url:         "/api/v1/users/" + c.user.Nickname + "/subscriptions/" + uuid,
			Data:        payload,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		output := &common.Response{}

		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		devs := c.devices
		newDevs := []models.Device{}
		for _, dev := range devs {
			if dev.UUID == uuid {
				continue
			}
			newDevs = append(newDevs, dev)
		}

		toast.Text(common.MSG_UNSUBSCRIBED_SUCCESS).Type(common.TTYPE_SUCCESS).Dispatch()

		ctx.Dispatch(func(ctx app.Context) {
			if uuid == c.thisDeviceUUID {
				c.subscribed = false
			}

			c.subscription.Mentions = false
			c.subscription.Replies = false

			c.thisDevice = models.Device{}
			c.deleteSubscriptionModalShow = false
			c.devices = newDevs
		})
	})
}

func (c *Content) handleUserDelete(ctx app.Context, a app.Action) {
	ctx.Dispatch(func(ctx app.Context) {
		c.settingsButtonDisabled = true
	})

	// Instantiate the toast.
	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		defer ctx.Dispatch(func(ctx app.Context) {
			c.settingsButtonDisabled = false
		})

		input := &common.CallInput{
			Method:      "DELETE",
			Url:         "/api/v1/users/" + c.user.Nickname,
			Data:        c.user,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		output := &common.Response{}

		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		// Invalidate the LocalStorage contents.
		_ = ctx.LocalStorage().Set("authGranted", false)
		if err := common.SaveUser(&models.User{}, &ctx); err != nil {
			toast.Text(common.ErrLocalStorageUserSave).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		ctx.Navigate("/logout")
	})
}
