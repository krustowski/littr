package settings

import (
	"crypto/sha512"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/models"

	//"github.com/SherClockHolmes/webpush-go"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

// onKeyDown()
func (c *Content) onKeyDown(ctx app.Context, e app.Event) {
	if e.Get("key").String() == "Escape" || e.Get("key").String() == "Esc" {
		ctx.NewAction("dismiss")
		return
	}
}

// onClickPass()
func (c *Content) onClickPass(ctx app.Context, e app.Event) {
	toast := common.Toast{AppContext: &ctx}

	c.settingsButtonDisabled = true

	defer ctx.Dispatch(func(ctx app.Context) {
		c.settingsButtonDisabled = false
	})

	ctx.Async(func() {
		// trim the padding spaces on the extremities
		// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
		passphrase := strings.TrimSpace(c.passphrase)
		passphraseAgain := strings.TrimSpace(c.passphraseAgain)
		passphraseCurrent := strings.TrimSpace(c.passphraseCurrent)

		if passphrase == "" || passphraseAgain == "" || passphraseCurrent == "" {
			toast.Text(common.ERR_PASSPHRASE_MISSING).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		if passphrase != passphraseAgain {
			toast.Text(common.ERR_PASSPHRASE_MISMATCH).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		//passHash := sha512.Sum512([]byte(passphrase + app.Getenv("APP_PEPPER")))
		passHash := sha512.Sum512([]byte(passphrase + common.AppPepper))
		passHashCurrent := sha512.Sum512([]byte(passphraseCurrent + common.AppPepper))

		payload := struct {
			NewPassphraseHex     string `json:"new_passphrase_hex"`
			CurrentPassphraseHex string `json:"current_passphrase_hex"`
		}{
			NewPassphraseHex:     fmt.Sprintf("%x", passHash),
			CurrentPassphraseHex: fmt.Sprintf("%x", passHashCurrent),
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
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		c.user.Passphrase = string(passHash[:])

		/*var userStream []byte
		if err := reload(c.user, &userStream); err != nil {
			toast.Text("cannot update local storage").Type("error").Dispatch(c, dispatch)

			ctx.Dispatch(func(ctx app.Context) {
				c.settingsButtonDisabled = false
			})
			return
		}*/

		toast.Text(common.MSG_PASSPHRASE_UPDATED).Type(common.TTYPE_SUCCESS).Dispatch(c, dispatch)
		return
	})
}

// onClickAbout()
func (c *Content) onClickAbout(ctx app.Context, e app.Event) {
	toast := common.Toast{AppContext: &ctx}

	c.settingsButtonDisabled = true

	defer ctx.Dispatch(func(ctx app.Context) {
		c.settingsButtonDisabled = false
	})

	ctx.Async(func() {
		// trim the padding spaces on the extremities
		// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
		aboutText := strings.TrimSpace(c.aboutText)

		if aboutText == "" {
			toast.Text(common.ERR_ABOUT_TEXT_UNCHANGED).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		if len(aboutText) > 100 {
			toast.Text(common.ERR_ABOUT_TEXT_CHAR_LIMIT).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		// see options.go
		payload := c.prefillPayload()
		payload.AboutText = aboutText

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
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		// Update the LocalStorage.
		common.SaveUser(&c.user, &ctx)

		toast.Text(common.MSG_ABOUT_TEXT_UPDATED).Type(common.TTYPE_SUCCESS).Dispatch(c, dispatch)
		return
	})
}

// onClickWebsite()
func (c *Content) onClickWebsite(ctx app.Context, e app.Event) {
	toast := common.Toast{AppContext: &ctx}

	c.settingsButtonDisabled = true

	defer ctx.Dispatch(func(ctx app.Context) {
		c.settingsButtonDisabled = false
	})

	ctx.Async(func() {
		website := strings.TrimSpace(c.website)

		// check the trimmed version of website string
		if website == "" {
			toast.Text(common.ERR_WEBSITE_UNCHANGED).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		// check the URL/URI format
		if _, err := url.ParseRequestURI(website); err != nil {
			toast.Text(common.ERR_WEBSITE_INVALID).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		// create a regex object
		regex, err := regexp.Compile("^(http|https)://")
		if err != nil {
			toast.Text(common.ERR_WEBSITE_REGEXP_FAIL).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		if !regex.MatchString(website) {
			website = "https://" + website
		}

		// see options.go
		payload := c.prefillPayload()
		payload.WebsiteLink = website

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
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		toast.Text(common.MSG_WEBSITE_UPDATED).Type(common.TTYPE_SUCCESS).Dispatch(c, dispatch)

		// Update the LocalStorage.
		common.SaveUser(&c.user, &ctx)

		ctx.Dispatch(func(ctx app.Context) {
			// update user's struct in memory
			c.user.Web = c.website
		})
		return
	})
	return
}

// onClickDeleteSubscription()
func (c *Content) onClickDeleteSubscription(ctx app.Context, e app.Event) {
	toast := common.Toast{AppContext: &ctx}

	c.settingsButtonDisabled = true

	defer ctx.Dispatch(func(ctx app.Context) {
		c.settingsButtonDisabled = false
	})

	uuid := c.interactedUUID
	if uuid == "" {
		toast.Text(common.ERR_SUBSCRIPTION_BLANK_UUID).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
		return
	}

	ctx.Async(func() {
		payload := struct {
			UUID string `json:"device_uuid"`
		}{
			UUID: uuid,
		}

		input := &common.CallInput{
			Method:      "DELETE",
			Url:         "/api/v1/push/subscription/" + ctx.DeviceID(),
			Data:        payload,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		output := &common.Response{}

		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
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

		toast.Text(common.MSG_UNSUBSCRIBED_SUCCESS).Type(common.TTYPE_SUCCESS).Dispatch(c, dispatch)

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
		return
	})
	return
}

// onDarkModeSwitch()
func (c *Content) onDarkModeSwitch(ctx app.Context, e app.Event) {
	//m := ctx.JSSrc().Get("checked").Bool()

	ctx.Dispatch(func(ctx app.Context) {
		c.darkModeOn = !c.darkModeOn

		ctx.LocalStorage().Set("mode", "dark")
		if !c.darkModeOn {
			ctx.LocalStorage().Set("mode", "light")
		}
	})

	app.Window().Get("LIT").Call("toggleMode")
	//c.app.Window().Get("body").Call("toggleClass", "lightmode")
}

// onCliclDeleteSubscriptionModalShow()
func (c *Content) onClickDeleteSubscriptionModalShow(ctx app.Context, e app.Event) {
	uuid := ctx.JSSrc().Get("id").String()

	ctx.Dispatch(func(ctx app.Context) {
		c.deleteSubscriptionModalShow = true
		c.settingsButtonDisabled = true
		c.interactedUUID = uuid
	})
}

// onClickDeleteAccountModalShow()
func (c *Content) onClickDeleteAccountModalShow(ctx app.Context, e app.Event) {
	ctx.Dispatch(func(ctx app.Context) {
		c.deleteAccountModalShow = true
		c.settingsButtonDisabled = true
	})
}

// onDismissToast
func (c *Content) onDismissToast(ctx app.Context, e app.Event) {
	ctx.NewAction("dismiss")
}

// onClickNotifSwitch()
func (c *Content) onClickNotifSwitch(ctx app.Context, e app.Event) {
	tag := ""
	source := ctx.JSSrc().Get("id").String()

	c.settingsButtonDisabled = true

	defer ctx.Dispatch(func(ctx app.Context) {
		c.settingsButtonDisabled = false
	})

	if strings.Contains(source, "reply") {
		tag = "reply"
	} else if strings.Contains(source, "mention") {
		tag = "mention"
	}

	checked := ctx.JSSrc().Get("checked").Bool()

	toast := common.Toast{AppContext: &ctx}

	// unsubscribe
	if !checked {
		// there's only one tag left, and user unchecked that particular one
		if len(c.thisDevice.Tags) == 1 && len(c.checkTags(c.thisDevice.Tags, tag)) == 0 {
			// DELETE
			c.deleteSubscription(ctx, tag)
			return
		}

		// otherwise we just update the current device with another tag (add/remove)
		// PUT
		c.updateSubscriptionTag(ctx, tag)
		return
	}

	// switched on --- add new device or update the existing one
	if !c.checkPermission(ctx, checked) {
		return
	}

	subscribed := false
	if c.subscription.Replies || c.subscription.Mentions {
		subscribed = true
		c.updateSubscriptionTag(ctx, tag)
		return
	}

	// add new device
	ctx.Async(func() {
		var deviceSub models.Device

		// fetch the unique identificator for notifications and unsubscribe option
		uuid := ctx.DeviceID()

		vapidPubKey := c.VAPIDpublic
		if vapidPubKey == "" {
			// compiled into frontend/common/api.VapidPublicKey variable in wasm binary (see Dockerfile for more info)
			vapidPubKey = common.VapidPublicKey
		}

		// register the subscription
		sub, err := ctx.Notifications().Subscribe(vapidPubKey)
		if err != nil {
			toast.Text(common.ERR_SUBSCRIPTION_REQ_FAIL+err.Error()).Type(common.TTYPE_ERR).Dispatch(c, dispatch)

			ctx.Dispatch(func(ctx app.Context) {
				c.settingsButtonDisabled = false
				//c.subscribed = false
			})
			return
		}

		// we need to convert type app.NotificationSubscription into webpush.Subscription
		webSub := models.Subscription{
			Endpoint: sub.Endpoint,
			Keys: models.Keys{
				Auth:   sub.Keys.Auth,
				P256dh: sub.Keys.P256dh,
			},
		}

		// compose the Device struct to be saved to the database
		deviceSub = models.Device{
			UUID:         uuid,
			TimeCreated:  time.Now(),
			Subscription: webSub,
			Tags: []string{
				tag,
			},
		}

		// send the registration to backend
		input := &common.CallInput{
			Method:      "POST",
			Url:         "/api/v1/push/subscription",
			Data:        deviceSub,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		output := &common.Response{}

		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch(c, dispatch)

			ctx.Dispatch(func(ctx app.Context) {
				c.subscribed = false
			})
			return
		}

		if output.Code != 201 && output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		devs := c.devices
		if !subscribed {
			devs = append(devs, deviceSub)
		}

		toast.Text(common.MSG_SUBSCRIPTION_REQ_SUCCESS).Type(common.TTYPE_SUCCESS).Dispatch(c, dispatch)

		// Update the LocalStorage.
		common.SaveUser(&c.user, &ctx)

		// dispatch the good news to client
		ctx.Dispatch(func(ctx app.Context) {
			//c.user = user
			c.subscribed = true

			if tag == "mention" {
				c.subscription.Mentions = !c.subscription.Mentions
			} else if tag == "reply" {
				c.subscription.Replies = !c.subscription.Replies
			}

			c.thisDevice = deviceSub
			c.devices = devs
		})
		return
	})
	return
}

// onLocalTimeModeSwitch()
func (c *Content) onLocalTimeModeSwitch(ctx app.Context, e app.Event) {
	c.settingsButtonDisabled = true

	defer ctx.Dispatch(func(ctx app.Context) {
		c.settingsButtonDisabled = false
	})

	toast := common.Toast{AppContext: &ctx}
	localTime := c.user.LocalTimeMode

	ctx.Async(func() {
		// see options.go
		payload := c.prefillPayload()
		payload.LocalTimeMode = !localTime

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
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch(c, dispatch)

			ctx.Dispatch(func(ctx app.Context) {
				c.user.LocalTimeMode = localTime
			})
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		toast.Text(common.MSG_LOCAL_TIME_TOGGLE).Type(common.TTYPE_SUCCESS).Dispatch(c, dispatch)

		// Update the LocalStorage.
		common.SaveUser(&c.user, &ctx)

		// dispatch the good news to client
		ctx.Dispatch(func(ctx app.Context) {
			c.user.LocalTimeMode = !localTime
		})
		return
	})
}

// onClickPrivateSwitch()
func (c *Content) onClickPrivateSwitch(ctx app.Context, e app.Event) {
	c.settingsButtonDisabled = true

	defer ctx.Dispatch(func(ctx app.Context) {
		c.settingsButtonDisabled = false
	})

	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		// see options.go
		payload := c.prefillPayload()
		payload.Private = !c.user.Private

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
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		toast.Text(common.MSG_PRIVATE_MODE_TOGGLE).Type(common.TTYPE_SUCCESS).Dispatch(c, dispatch)

		// Update the LocalStorage.
		common.SaveUser(&c.user, &ctx)

		// dispatch the good news to client
		ctx.Dispatch(func(ctx app.Context) {
			c.user.Private = !c.user.Private
		})
		return
	})
}

// onClickDeleteAccount()
func (c *Content) onClickDeleteAccount(ctx app.Context, e app.Event) {
	// Instantiate the toast.
	toast := common.Toast{AppContext: &ctx}

	c.settingsButtonDisabled = true

	defer ctx.Dispatch(func(ctx app.Context) {
		c.settingsButtonDisabled = false
	})

	// Invalidate the LocalStorage contents.
	ctx.LocalStorage().Set("authGranted", false)
	common.SaveUser(&models.User{}, &ctx)

	ctx.Async(func() {
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
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		//c.userLogged = false
		//ctx.Navigate("/login")
		ctx.Navigate("/logout")
	})
	return
}

// handleFigUpload() --> common/image.go TODO
func (c *Content) handleFigUpload(ctx app.Context, e app.Event) {
	toast := common.Toast{AppContext: &ctx}

	file := e.Get("target").Get("files").Index(0)

	c.settingsButtonDisabled = true

	defer ctx.Dispatch(func(ctx app.Context) {
		c.settingsButtonDisabled = false
	})

	ctx.Async(func() {
		if figData, err := common.ReadFile(file); err != nil {
			toast.Text(err.Error()).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return

		} else {
			/*payload := models.Post{
				Nickname:  author,
				Type:      "fig",
				Content:   file.Get("name").String(),
				Timestamp: time.Now(),
				Data:      data,
			}*/

			// add new post/poll to backend struct
			/*if _, ok := littrAPI("POST", path, payload, user.Nickname, 0); !ok {
				toastText = "backend error: cannot add new content"
				log.Println("cannot post new content to API!")
			} else {
				ctx.Navigate("/flow")
			}*/

			path := "/api/v1/users/" + c.user.Nickname + "/avatar"

			payload := models.Post{
				Nickname: c.user.Nickname,
				Figure:   file.Get("name").String(),
				Data:     figData,
			}

			input := &common.CallInput{
				Method:      "POST",
				Url:         path,
				Data:        payload,
				CallerID:    c.user.Nickname,
				PageNo:      0,
				HideReplies: false,
			}

			type dataModel struct {
				Key string
			}

			output := &common.Response{Data: &dataModel{}}

			if ok := common.FetchData(input, output); !ok {
				toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
				return
			}

			if output.Code != 200 {
				toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
				return
			}

			data, ok := output.Data.(*dataModel)
			if !ok {
				toast.Text(common.ERR_CANNOT_GET_DATA).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
				return
			}

			avatar := "/web/pix/thumb_" + data.Key

			toast.Text(common.MSG_AVATAR_CHANGE_SUCCESS).Type(common.TTYPE_SUCCESS).Dispatch(c, dispatch)

			// Update the LocalStorage.
			common.SaveUser(&c.user, &ctx)

			ctx.Dispatch(func(ctx app.Context) {
				c.newFigFile = file.Get("name").String()
				c.newFigData = figData

				c.user.AvatarURL = avatar
			})
			return
		}
	})
}
