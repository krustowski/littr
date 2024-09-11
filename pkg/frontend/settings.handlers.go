package frontend

import (
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"

	"go.vxn.dev/littr/pkg/helpers"
	"go.vxn.dev/littr/pkg/models"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type OptionsPayload struct {
	UIDarkMode    bool   `json:"dark_mode"`
	LiveMode      bool   `json:"live_mode"`
	LocalTimeMode bool   `json:"local_time_mode"`
	Private       bool   `json:"private"`
	AboutText     string `json:"about_you"`
	WebsiteLink   string `json:"website_link"`
}

func (c *settingsContent) prefillPayload() OptionsPayload {

	payload := OptionsPayload{
		UIDarkMode:    c.user.UIDarkMode,
		LiveMode:      c.user.LiveMode,
		LocalTimeMode: c.user.LocalTimeMode,
		Private:       c.user.Private,
		AboutText:     c.user.About,
		WebsiteLink:   c.user.Web,
	}

	return payload
}

func (c *settingsContent) handleDismiss(ctx app.Context, a app.Action) {
	ctx.Dispatch(func(ctx app.Context) {
		c.toastText = ""
		c.toastType = ""
		c.toastShow = (c.toastText != "")
		c.settingsButtonDisabled = false
		c.deleteAccountModalShow = false
		c.deleteSubscriptionModalShow = false
	})
}

func (c *settingsContent) onKeyDown(ctx app.Context, e app.Event) {
	if e.Get("key").String() == "Escape" || e.Get("key").String() == "Esc" {
		ctx.NewAction("dismiss")
		return
	}
}

// onClickPass
func (c *settingsContent) onClickPass(ctx app.Context, e app.Event) {
	toastText := ""

	c.settingsButtonDisabled = true

	ctx.Async(func() {
		// trim the padding spaces on the extremities
		// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
		passphrase := strings.TrimSpace(c.passphrase)
		passphraseAgain := strings.TrimSpace(c.passphraseAgain)
		passphraseCurrent := strings.TrimSpace(c.passphraseCurrent)

		if passphrase == "" || passphraseAgain == "" || passphraseCurrent == "" {
			toastText = "passphrase fields need to be filled"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.settingsButtonDisabled = false
			})
			return
		}

		if passphrase != passphraseAgain {
			toastText = "passphrases do not match"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.settingsButtonDisabled = false
			})
			return
		}

		//passHash := sha512.Sum512([]byte(passphrase + app.Getenv("APP_PEPPER")))
		passHash := sha512.Sum512([]byte(passphrase + appPepper))
		passHashCurrent := sha512.Sum512([]byte(passphraseCurrent + appPepper))

		payload := struct {
			NewPassphraseHex     string `json:"new_passphrase_hex"`
			CurrentPassphraseHex string `json:"current_passphrase_hex"`
		}{
			NewPassphraseHex:     fmt.Sprintf("%x", passHash),
			CurrentPassphraseHex: fmt.Sprintf("%x", passHashCurrent),
		}

		response := struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		}{}

		input := callInput{
			Method: "PATCH",
			Url: "/api/v1/users/"+c.user.Nickname+"/passphrase",
			Data: payload,
			CallerID: c.user.Nickname,
			PageNo: 0,
			HideReplies: false,
		}

		if data, ok := littrAPI(input); ok {
			if err := json.Unmarshal(*data, &response); err != nil {
				toastText = "JSON parse error: " + err.Error()

				ctx.Dispatch(func(ctx app.Context) {
					c.toastText = toastText
					c.toastShow = (toastText != "")
					c.settingsButtonDisabled = false
				})
				return
			}
		}

		log.Println(response.Code)

		if response.Code != 200 {
			toastText = response.Message
			//toastText = "passphrase updating error, try again later"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.settingsButtonDisabled = false
			})
			return
		}

		c.user.Passphrase = string(passHash[:])

		var userStream []byte
		if err := reload(c.user, &userStream); err != nil {
			toastText = "cannot update local storage"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.settingsButtonDisabled = false
			})
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.toastType = "success"
			c.toastText = "passphrase updated"
			c.toastShow = (toastText != "")
			c.settingsButtonDisabled = false
		})
		return
	})
}

// onClickAbout
func (c *settingsContent) onClickAbout(ctx app.Context, e app.Event) {
	toastText := ""

	c.settingsButtonDisabled = true

	ctx.Async(func() {
		// trim the padding spaces on the extremities
		// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
		aboutText := strings.TrimSpace(c.aboutText)

		if aboutText == "" {
			toastText = "about textarea needs to be filled, or you prolly haven't changed the text"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.settingsButtonDisabled = false
			})
			return
		}

		if len(aboutText) > 100 {
			toastText = "about text has to be shorter than 100 chars"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.settingsButtonDisabled = false
			})
			return
		}

		payload := c.prefillPayload()
		payload.AboutText = aboutText

		input := callInput{
			Method: "PATCH",
			Url: "/api/v1/users/"+c.user.Nickname+"/options",
			Data: payload,
			CallerID: c.user.Nickname,
			PageNo: 0,
			HideReplies: false,
		}

		if _, ok := littrAPI(input); !ok {
			toastText = "generic backend error"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.settingsButtonDisabled = false
			})
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.toastText = "about text updated"
			c.toastShow = (toastText != "")
			c.toastType = "success"
			c.settingsButtonDisabled = false
		})
		return
	})
}

func (c *settingsContent) onClickWebsite(ctx app.Context, e app.Event) {
	toastText := ""

	c.settingsButtonDisabled = true

	ctx.Async(func() {
		website := strings.TrimSpace(c.website)

		// check the trimmed version of website string
		if website == "" {
			toastText = "website URL has to be filled, or changed"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.settingsButtonDisabled = false
			})
			return
		}

		// check the URL/URI format
		if _, err := url.ParseRequestURI(website); err != nil {
			toastText = "website prolly not a valid URL"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.settingsButtonDisabled = false
			})
			return
		}

		// create a regex object
		regex, err := regexp.Compile("^(http|https)://")
		if err != nil {
			toastText := "failed to check the website (regex object fail)"
			log.Println(toastText)

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.settingsButtonDisabled = false
			})
			return
		}

		if !regex.MatchString(website) {
			website = "https://" + website
		}

		payload := c.prefillPayload()
		payload.WebsiteLink = website

		input := callInput{
			Method: "PATCH",
			Url: "/api/v1/users/"+c.user.Nickname+"/options",
			Data: payload,
			CallerID: c.user.Nickname,
			PageNo: 0,
			HideReplies: false,
		}

		if _, ok := littrAPI(input); !ok {
			toastText = "generic backend error"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.settingsButtonDisabled = false
			})
			return
		}

		c.user.Web = c.website

		//ctx.Navigate("/users")
		ctx.Dispatch(func(ctx app.Context) {
			c.toastText = "website updated"
			c.toastShow = (toastText != "")
			c.toastType = "success"
			c.settingsButtonDisabled = false
		})
		return
	})
	return
}

func (c *settingsContent) onClickDeleteSubscription(ctx app.Context, e app.Event) {
	toastText := ""

	c.settingsButtonDisabled = true

	uuid := c.interactedUUID
	if uuid == "" {
		toastText := "blank UUID string"

		ctx.Dispatch(func(ctx app.Context) {
			c.toastText = toastText
			c.toastShow = toastText != ""
			c.settingsButtonDisabled = false
		})
		return
	}

	payload := struct {
		UUID string `json:"device_uuid"`
	}{
		UUID: uuid,
	}

	ctx.Async(func() {
		input := callInput{
			Method: "DELETE",
			Url: "/api/v1/push/subscription/"+ctx.DeviceID(),
			Data: payload,
			CallerID: c.user.Nickname,
			PageNo: 0,
			HideReplies: false,
		}

		if _, ok := littrAPI(input); !ok {
			ctx.Dispatch(func(ctx app.Context) {
				//c.toastText = toastText
				c.toastText = "failed to unsubscribe, try again later"
				c.toastShow = toastText != ""
				c.settingsButtonDisabled = false
			})
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

		toastText = "device successfully unsubscribed"

		ctx.Dispatch(func(ctx app.Context) {
			c.toastText = toastText
			c.toastShow = toastText != ""
			c.toastType = "info"
			c.settingsButtonDisabled = false

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

func (c *settingsContent) onDarkModeSwitch(ctx app.Context, e app.Event) {
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

func (c *settingsContent) onClickDeleteSubscriptionModalShow(ctx app.Context, e app.Event) {
	uuid := ctx.JSSrc().Get("id").String()

	ctx.Dispatch(func(ctx app.Context) {
		c.deleteSubscriptionModalShow = true
		c.settingsButtonDisabled = true
		c.interactedUUID = uuid
	})
}

func (c *settingsContent) onClickDeleteAccountModalShow(ctx app.Context, e app.Event) {
	ctx.Dispatch(func(ctx app.Context) {
		c.deleteAccountModalShow = true
		c.settingsButtonDisabled = true
	})
}

func (c *settingsContent) dismissToast(ctx app.Context, e app.Event) {
	ctx.NewAction("dismiss")
}

func (c *settingsContent) checkTags(tags []string, tag string) []string {
	// delete the tag if tags contain it
	if helpers.Contains(tags, tag) {
		newTags := []string{}
		for _, t := range tags {
			if t == tag {
				continue
			}
			newTags = append(newTags, t)
		}
		return newTags
	}

	// add the tag if missing
	return append(tags, tag)
}

func (c *settingsContent) deleteSubscription(ctx app.Context, tag string) {
	toastText := ""
	uuid := ctx.DeviceID()

	c.settingsButtonDisabled = true

	payload := struct {
		UUID string `json:"device_uuid"`
	}{
		UUID: uuid,
	}

	devs := c.devices
	newDevs := []models.Device{}
	for _, dev := range devs {
		if dev.UUID == ctx.DeviceID() {
			continue
		}
		newDevs = append(newDevs, dev)
	}

	ctx.Async(func() {
		input := callInput{
			Method: "DELETE",
			Url: "/api/v1/push/subscription/"+ctx.DeviceID(),
			Data: payload,
			CallerID: c.user.Nickname,
			PageNo: 0,
			HideReplies: false,
		}

		if _, ok := littrAPI(input); !ok {
			ctx.Dispatch(func(ctx app.Context) {
				//c.toastText = toastText
				c.toastText = "failed to unsubscribe, try again later"
				c.toastShow = toastText != ""

				c.subscribed = true
				c.settingsButtonDisabled = false
			})
			return

		}

		ctx.Dispatch(func(ctx app.Context) {
			//c.toastText = toastText
			c.toastText = "successfully unsubscribed, notifications off"
			c.toastShow = toastText != ""
			c.toastType = "info"

			c.settingsButtonDisabled = false

			c.subscription.Mentions = false
			c.subscription.Replies = false

			c.subscribed = false
			c.thisDevice = models.Device{}
			c.devices = newDevs
		})
		return
	})
	return
}

func (c *settingsContent) checkPermission(ctx app.Context, checked bool) bool {
	toastText := ""

	// notify user that their browser does not support notifications and therefore they cannot
	// use notifications service
	if c.notificationPermission == app.NotificationNotSupported && checked {
		toastText = "notifications are not supported in this browser"

		ctx.Dispatch(func(ctx app.Context) {
			c.toastText = toastText
			c.toastShow = (toastText != "")
			c.toastType = "error"

			c.replyNotifOn = false
			c.subscribed = false
		})
		return false
	}

	// request the permission on default when switch is toggled
	if (c.notificationPermission == app.NotificationDefault && checked) ||
		(c.notificationPermission == app.NotificationDenied) {
		c.notificationPermission = ctx.Notifications().RequestPermission()
	}

	// fail on denied permission
	if c.notificationPermission != app.NotificationGranted {
		toastText = "notification permission denied by user"

		ctx.Dispatch(func(ctx app.Context) {
			c.toastText = toastText
			c.toastShow = (toastText != "")

			c.replyNotifOn = false
			c.subscribed = false
		})
		return false
	}
	return true
}

func (c *settingsContent) updateSubscriptionTag(ctx app.Context, tag string) {
	c.settingsButtonDisabled = true

	devs := c.devices
	newDevs := []models.Device{}
	for _, dev := range devs {
		if dev.UUID == ctx.DeviceID() {
			if len(c.checkTags(dev.Tags, tag)) == 0 {
				continue
			}
			dev.Tags = c.checkTags(dev.Tags, tag)
		}
		newDevs = append(newDevs, dev)
	}

	deviceSub := c.thisDevice

	ctx.Async(func() {
		input := callInput{
			Method: "PUT",
			Url: "/api/v1/push/subscription/"+ctx.DeviceID()+"/"+tag,
			Data: deviceSub,
			CallerID: c.user.Nickname,
			PageNo: 0,
			HideReplies: false,
		}

		if _, ok := littrAPI(input); !ok {
			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = "failed to update the subscription, try again later"
				c.toastShow = c.toastText != ""

				//c.subscribed = true
				c.settingsButtonDisabled = false
			})
			return

		}

		deviceSub.Tags = c.checkTags(c.thisDevice.Tags, tag)

		ctx.Dispatch(func(ctx app.Context) {
			c.toastText = "subscription updated"
			c.toastShow = c.toastText != ""
			c.toastType = "info"

			if tag == "mention" {
				c.subscription.Mentions = !c.subscription.Mentions
			} else if tag == "reply" {
				c.subscription.Replies = !c.subscription.Replies
			}

			c.thisDevice = deviceSub
			c.devices = newDevs
			c.settingsButtonDisabled = false
		})
		return
	})
	return
}

func (c *settingsContent) onClickNotifSwitch(ctx app.Context, e app.Event) {
	tag := ""
	source := ctx.JSSrc().Get("id").String()

	c.settingsButtonDisabled = true

	if strings.Contains(source, "reply") {
		tag = "reply"
	} else if strings.Contains(source, "mention") {
		tag = "mention"
	}

	checked := ctx.JSSrc().Get("checked").Bool()
	toastText := ""

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
			// compiled into frontend/helpers.vapidPublicKey variable in wasm binary (see Dockerfile for more info)
			vapidPubKey = vapidPublicKey
		}

		// register the subscription
		sub, err := ctx.Notifications().Subscribe(vapidPubKey)
		if err != nil {
			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = "failed to subscribe to notifications: " + err.Error()
				c.toastShow = c.toastText != ""
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
		input := callInput{
			Method: "POST",
			Url: "/api/v1/push/subscription",
			Data: deviceSub,
			CallerID: c.user.Nickname,
			PageNo: 0,
			HideReplies: false,
		}

		if _, ok := littrAPI(input); !ok {
			toastText := "cannot reach backend!"

			ctx.Dispatch(func(ctx app.Context) {
				//c.toastText = toastText
				c.toastText = "failed to subscribe to notifications"
				c.toastShow = toastText != ""
				c.settingsButtonDisabled = false

				c.subscribed = false
			})
			return
		}

		devs := c.devices
		if !subscribed {
			devs = append(devs, deviceSub)
		}

		// dispatch the good news to client
		ctx.Dispatch(func(ctx app.Context) {
			//c.user = user
			c.subscribed = true

			c.toastText = "successfully subscribed"
			c.toastShow = toastText != ""
			c.toastType = "success"
			c.settingsButtonDisabled = false

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

func (c *settingsContent) onLocalTimeModeSwitch(ctx app.Context, e app.Event) {
	c.settingsButtonDisabled = true

	toastText := ""
	localTime := c.user.LocalTimeMode

	ctx.Async(func() {

		payload := c.prefillPayload()
		payload.LocalTimeMode = !localTime

		input := callInput{
			Method: "PATCH",
			Url: "/api/v1/users/"+c.user.Nickname+"/options",
			Data: payload,
			CallerID: c.user.Nickname,
			PageNo: 0,
			HideReplies: false,
		}

		if _, ok := littrAPI(input); !ok {
			toastText = "cannot reach backend!"

			ctx.Dispatch(func(ctx app.Context) {
				//c.toastText = toastText
				c.toastText = "failed to toggle the local time mode"
				c.toastShow = toastText != ""
				c.settingsButtonDisabled = false

				c.user.LocalTimeMode = localTime
			})
			return
		}

		// dispatch the good news to client
		ctx.Dispatch(func(ctx app.Context) {
			c.toastText = "local time mode toggled"
			c.toastShow = toastText != ""
			c.toastType = "success"
			c.settingsButtonDisabled = false

			c.user.LocalTimeMode = !localTime
		})
		return
	})
}

func (c *settingsContent) onClickPrivateSwitch(ctx app.Context, e app.Event) {
	c.settingsButtonDisabled = true

	toastText := ""

	ctx.Async(func() {

		payload := c.prefillPayload()
		payload.Private = !c.user.Private

		input := callInput{
			Method: "PATCH",
			Url: "/api/v1/users/"+c.user.Nickname+"/options",
			Data: payload,
			CallerID: c.user.Nickname,
			PageNo: 0,
			HideReplies: false,
		}

		if _, ok := littrAPI(input); !ok {
			toastText = "cannot reach backend!"

			ctx.Dispatch(func(ctx app.Context) {
				//c.toastText = toastText
				c.toastText = "failed to toggle the private mode"
				c.toastShow = toastText != ""
				c.settingsButtonDisabled = false
			})
			return
		}

		// dispatch the good news to client
		ctx.Dispatch(func(ctx app.Context) {
			c.toastText = "private mode toggled"
			c.toastShow = toastText != ""
			c.toastType = "success"
			c.settingsButtonDisabled = false

			c.user.Private = !c.user.Private
		})
		return
	})
}

func (c *settingsContent) onClickDeleteAccount(ctx app.Context, e app.Event) {
	toastText := ""

	c.settingsButtonDisabled = true

	//ctx.LocalStorage().Set("userLogged", false)
	ctx.LocalStorage().Set("user", "")

	ctx.Async(func() {
		input := callInput{
			Method: "DELETE",
			Url: "/api/v1/users/"+c.user.Nickname,
			Data: c.user,
			CallerID: c.user.Nickname,
			PageNo: 0,
			HideReplies: false,
		}

		if _, ok := littrAPI(input); !ok {
			toastText = "generic backend error"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.settingsButtonDisabled = false
			})
			return
		}

		//c.userLogged = false
		//ctx.Navigate("/login")
		ctx.Navigate("/logout")
	})
	return
}

func (c *settingsContent) handleFigUpload(ctx app.Context, e app.Event) {
	var toastText string

	file := e.Get("target").Get("files").Index(0)

	c.settingsButtonDisabled = true

	ctx.Async(func() {
		if data, err := readFile(file); err != nil {
			toastText = err.Error()

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.settingsButtonDisabled = false
			})
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
				Data:     data,
			}

			resp := struct {
				Key string
			}{}

		input := callInput{
			Method: "POST",
			Url: path,
			Data: payload,
			CallerID: c.user.Nickname,
			PageNo: 0,
			HideReplies: false,
		}

			if raw, ok := littrAPI(input); ok {
				if err := json.Unmarshal(*raw, &resp); err != nil {
					toastText = "JSON parse error: " + err.Error()
					ctx.Dispatch(func(ctx app.Context) {
						c.toastText = toastText
						c.toastShow = (toastText != "")
						c.settingsButtonDisabled = false
					})
					return
				}

			} else {
				//ctx.Navigate("/flow")
				toastText = "generic backend error: cannot process the request"

				ctx.Dispatch(func(ctx app.Context) {
					c.toastText = toastText
					c.toastShow = (toastText != "")
					c.settingsButtonDisabled = false
				})
				return
			}

			toastText = "avatar successfully updated"

			avatar := "/web/pix/thumb_" + resp.Key

			ctx.Dispatch(func(ctx app.Context) {
				c.toastType = "success"
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.settingsButtonDisabled = false

				c.newFigFile = file.Get("name").String()
				c.newFigData = data

				c.user.AvatarURL = avatar
			})
			return
		}
	})
}
