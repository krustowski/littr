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

	"github.com/SherClockHolmes/webpush-go"
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

	// nasty
	c.settingsButtonDisabled = true

	ctx.Async(func() {
		// trim the padding spaces on the extremities
		// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
		passphrase := strings.TrimSpace(c.passphrase)
		passphraseAgain := strings.TrimSpace(c.passphraseAgain)
		passphraseCurrent := strings.TrimSpace(c.passphraseCurrent)

		if passphrase == "" || passphraseAgain == "" || passphraseCurrent == "" {
			toast.Text("passphrase fields need to be filled").Type("error").Dispatch(c, dispatch)
			return
		}

		if passphrase != passphraseAgain {
			toast.Text("passphrases do not match").Type("error").Dispatch(c, dispatch)
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
			toast.Text("cannot reach backend").Type("error").Dispatch(c, dispatch)
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type("error").Dispatch(c, dispatch)
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

		toast.Text("passphrase updated").Type("success").Dispatch(c, dispatch)
		return
	})
}

// onClickAbout()
func (c *Content) onClickAbout(ctx app.Context, e app.Event) {
	toast := common.Toast{AppContext: &ctx}

	c.settingsButtonDisabled = true

	ctx.Async(func() {
		// trim the padding spaces on the extremities
		// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
		aboutText := strings.TrimSpace(c.aboutText)

		if aboutText == "" {
			toast.Text("about textarea needs to be filled, or you prolly haven't changed the text").Type("error").Dispatch(c, dispatch)
			return
		}

		if len(aboutText) > 100 {
			toast.Text("about text has to be shorter than 100 chars").Type("error").Dispatch(c, dispatch)
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
			toast.Text("generic backend error").Type("error").Dispatch(c, dispatch)
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type("error").Dispatch(c, dispatch)
			return
		}

		toast.Text("about text updated").Type("success").Dispatch(c, dispatch)
		return
	})
}

// onClickWebsite()
func (c *Content) onClickWebsite(ctx app.Context, e app.Event) {
	toast := common.Toast{AppContext: &ctx}

	c.settingsButtonDisabled = true

	ctx.Async(func() {
		website := strings.TrimSpace(c.website)

		// check the trimmed version of website string
		if website == "" {
			toast.Text("website URL has to be filled, or changed").Type("error").Dispatch(c, dispatch)
			return
		}

		// check the URL/URI format
		if _, err := url.ParseRequestURI(website); err != nil {
			toast.Text("website prolly not a valid URL").Type("error").Dispatch(c, dispatch)
			return
		}

		// create a regex object
		regex, err := regexp.Compile("^(http|https)://")
		if err != nil {
			toast.Text("failed to check the website (regex object fail)").Type("error").Dispatch(c, dispatch)
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
			toast.Text("generic backend error").Type("error").Dispatch(c, dispatch)
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type("error").Dispatch(c, dispatch)
			return
		}

		toast.Text("website updated").Type("success").Dispatch(c, dispatch)

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

	uuid := c.interactedUUID
	if uuid == "" {
		toast.Text("blank UUID string").Type("error").Dispatch(c, dispatch)
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
			toast.Text("failed to unsubscribe, try again later").Type("error").Dispatch(c, dispatch)
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type("error").Dispatch(c, dispatch)
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

		toast.Text("device successfully unsubscribed").Type("success").Dispatch(c, dispatch)

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
			toast.Text("failed to subscribe to notifications: "+err.Error()).Type("error").Dispatch(c, dispatch)

			ctx.Dispatch(func(ctx app.Context) {
				c.settingsButtonDisabled = false
				//c.subscribed = false
			})
			return
		}

		// we need to convert type app.NotificationSubscription into webpush.Subscription
		webSub := webpush.Subscription{
			Endpoint: sub.Endpoint,
			Keys: webpush.Keys{
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
			toast.Text("backend connection failed").Type("error").Dispatch(c, dispatch)

			ctx.Dispatch(func(ctx app.Context) {
				c.subscribed = false
			})
			return
		}

		if output.Code != 201 && output.Code != 200 {
			toast.Text(output.Message).Type("error").Dispatch(c, dispatch)
			return
		}

		devs := c.devices
		if !subscribed {
			devs = append(devs, deviceSub)
		}

		toast.Text("successfully subscribed to notifs").Type("success").Dispatch(c, dispatch)

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
			toast.Text("cannot reach backend!").Type("error").Dispatch(c, dispatch)

			ctx.Dispatch(func(ctx app.Context) {
				c.user.LocalTimeMode = localTime
			})
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type("error").Dispatch(c, dispatch)
			return
		}

		toast.Text("local time mode toggled").Type("success").Dispatch(c, dispatch)

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
			toast.Text("cannot reach backend!").Type("error").Dispatch(c, dispatch)
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type("error").Dispatch(c, dispatch)
			return
		}

		toast.Text("private mode toggled").Type("success").Dispatch(c, dispatch)

		// dispatch the good news to client
		ctx.Dispatch(func(ctx app.Context) {
			c.user.Private = !c.user.Private
		})
		return
	})
}

// onClickDeleteAccount()
func (c *Content) onClickDeleteAccount(ctx app.Context, e app.Event) {
	toast := common.Toast{AppContext: &ctx}

	c.settingsButtonDisabled = true

	//ctx.LocalStorage().Set("userLogged", false)
	ctx.LocalStorage().Set("user", "")

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
			toast.Text("generic backend error").Type("error").Dispatch(c, dispatch)
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type("error").Dispatch(c, dispatch)
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

	ctx.Async(func() {
		if figData, err := common.ReadFile(file); err != nil {
			toast.Text(err.Error()).Type("error").Dispatch(c, dispatch)
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
				toast.Text("cannot reach backend!").Type("error").Dispatch(c, dispatch)
				return
			}

			if output.Code != 200 {
				toast.Text(output.Message).Type("error").Dispatch(c, dispatch)
				return
			}

			data, ok := output.Data.(*dataModel)
			if !ok {
				toast.Text("cannot get data").Type("error").Dispatch(c, dispatch)
				return
			}

			avatar := "/web/pix/thumb_" + data.Key

			toast.Text("avatar successfully updated").Type("success").Dispatch(c, dispatch)

			ctx.Dispatch(func(ctx app.Context) {
				c.newFigFile = file.Get("name").String()
				c.newFigData = figData

				c.user.AvatarURL = avatar
			})
			return
		}
	})
}
