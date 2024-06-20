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

	"go.savla.dev/littr/pkg/helpers"
	"go.savla.dev/littr/pkg/models"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

// onClickPass
func (c *settingsContent) onClickPass(ctx app.Context, e app.Event) {
	toastText := ""

	c.settingsButtonDisabled = true

	ctx.Async(func() {
		// trim the padding spaces on the extremities
		// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
		passphrase := strings.TrimSpace(c.passphrase)
		passphraseAgain := strings.TrimSpace(c.passphraseAgain)

		if passphrase == "" || passphraseAgain == "" {
			toastText = "both passphrases need to be filled, or text changed"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		if passphrase != passphraseAgain {
			toastText = "passphrases do not match"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		//passHash := sha512.Sum512([]byte(passphrase + app.Getenv("APP_PEPPER")))
		passHash := sha512.Sum512([]byte(passphrase + appPepper))

		updatedUser := c.user
		//updatedUser.Passphrase = string(passHash[:])
		updatedUser.PassphraseHex = fmt.Sprintf("%x", passHash)

		response := struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		}{}

		if data, ok := litterAPI("PUT", "/api/v1/users/"+updatedUser.Nickname, updatedUser, c.user.Nickname, 0); !ok {
			if err := json.Unmarshal(*data, &response); err != nil {
				toastText = "JSON parse error: " + err.Error()
			}
			toastText = "generic backend error: " + response.Message

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
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
			})
			return
		}

		ctx.Navigate("/users")
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
			})
			return
		}

		if len(aboutText) > 100 {
			toastText = "about text has to be shorter than 100 chars"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		updatedUser := c.user
		updatedUser.About = aboutText

		if _, ok := litterAPI("PUT", "/api/v1/users/"+updatedUser.Nickname, updatedUser, c.user.Nickname, 0); !ok {
			toastText = "generic backend error"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		c.user.About = c.aboutText

		var userStream []byte
		if err := reload(c.user, &userStream); err != nil {
			toastText = "cannot update local storage"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		ctx.LocalStorage().Set("user", userStream)

		ctx.Navigate("/users")
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
			})
			return
		}

		// check the URL/URI format
		if _, err := url.ParseRequestURI(website); err != nil {
			toastText = "website prolly not a valid URL"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
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
			})
			return
		}

		if !regex.MatchString(website) {
			website = "https://" + website
		}

		updatedUser := c.user
		updatedUser.Web = website

		if _, ok := litterAPI("PUT", "/api/v1/users/"+updatedUser.Nickname, updatedUser, c.user.Nickname, 0); !ok {
			toastText = "generic backend error"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		c.user.Web = c.website

		var userStream []byte
		if err := reload(c.user, &userStream); err != nil {
			toastText = "cannot update local storage"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		ctx.Navigate("/users")
	})
	return
}

func (c *settingsContent) onClickDeleteSubscription(ctx app.Context, e app.Event) {
	toastText := ""

	uuid := c.interactedUUID
	if uuid == "" {
		toastText := "blank UUID string"

		ctx.Dispatch(func(ctx app.Context) {
			c.toastText = toastText
			c.toastShow = toastText != ""
		})
		return
	}

	payload := struct {
		UUID string `json:"device_uuid"`
	}{
		UUID: uuid,
	}

	ctx.Async(func() {
		if _, ok := litterAPI("DELETE", "/api/v1/push/subscription/"+ctx.DeviceID(), payload, c.user.Nickname, 0); !ok {
			ctx.Dispatch(func(ctx app.Context) {
				//c.toastText = toastText
				c.toastText = "failed to unsubscribe, try again later"
				c.toastShow = toastText != ""
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
	ctx.Dispatch(func(ctx app.Context) {
		c.toastText = ""
		c.toastType = ""
		c.toastShow = (c.toastText != "")
		c.settingsButtonDisabled = false
		c.deleteAccountModalShow = false
		c.deleteSubscriptionModalShow = false
	})
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
		if _, ok := litterAPI("DELETE", "/api/v1/push/subscription/"+ctx.DeviceID(), payload, c.user.Nickname, 0); !ok {
			ctx.Dispatch(func(ctx app.Context) {
				//c.toastText = toastText
				c.toastText = "failed to unsubscribe, try again later"
				c.toastShow = toastText != ""

				c.subscribed = true
			})
			return

		}

		ctx.Dispatch(func(ctx app.Context) {
			//c.toastText = toastText
			c.toastText = "successfully unsubscribed, notifications off"
			c.toastShow = toastText != ""
			c.toastType = "info"

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
		if _, ok := litterAPI("PUT", "/api/v1/push/subscription/"+ctx.DeviceID()+"/"+tag, deviceSub, c.user.Nickname, 0); !ok {
			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = "failed to update the subscription, try again later"
				c.toastShow = c.toastText != ""

				//c.subscribed = true
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
		})
		return
	})
	return
}

func (c *settingsContent) onClickNotifSwitch(ctx app.Context, e app.Event) {
	tag := ""
	source := ctx.JSSrc().Get("id").String()

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

		// send the registeration to backend
		if _, ok := litterAPI("POST", "/api/v1/push/subscription", deviceSub, c.user.Nickname, 0); !ok {
			toastText := "cannot reach backend!"

			ctx.Dispatch(func(ctx app.Context) {
				//c.toastText = toastText
				c.toastText = "failed to subscribe to notifications"
				c.toastShow = toastText != ""

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

func (c *settingsContent) onClickPrivateSwitch(ctx app.Context, e app.Event) {
	toastText := ""
	private := c.user.Private

	ctx.Async(func() {
		// send the registeration to backend
		if _, ok := litterAPI("PATCH", "/api/v1/users/"+c.user.Nickname+"/private", nil, c.user.Nickname, 0); !ok {
			toastText = "cannot reach backend!"

			ctx.Dispatch(func(ctx app.Context) {
				//c.toastText = toastText
				c.toastText = "failed to toggle the private mode"
				c.toastShow = toastText != ""

				c.user.Private = private
			})
			return
		}

		// dispatch the good news to client
		ctx.Dispatch(func(ctx app.Context) {
			c.toastText = "private mode toggled"
			c.toastShow = toastText != ""
			c.toastType = "success"

			c.user.Private = !private
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
		if _, ok := litterAPI("DELETE", "/api/v1/users/"+c.user.Nickname, c.user, c.user.Nickname, 0); !ok {
			toastText = "generic backend error"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
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

	//c.postButtonsDisabled = true

	ctx.Async(func() {
		if data, err := readFile(file); err != nil {
			toastText = err.Error()

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
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
			/*if _, ok := litterAPI("POST", path, payload, user.Nickname, 0); !ok {
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

			if raw, ok := litterAPI("POST", path, payload, c.user.Nickname, 0); ok {
				if err := json.Unmarshal(*raw, &resp); err != nil {
					toastText = "JSON parse error: " + err.Error()
					ctx.Dispatch(func(ctx app.Context) {
						c.toastText = toastText
						c.toastShow = (toastText != "")
					})
					return
				}

			} else {
				//ctx.Navigate("/flow")
				toastText = "generic backend error: cannot process the request"

				ctx.Dispatch(func(ctx app.Context) {
					c.toastText = toastText
					c.toastShow = (toastText != "")
				})
				return
			}

			toastText = "avatar successfully updated"

			avatar := "/web/pix/thumb_" + resp.Key

			ctx.Dispatch(func(ctx app.Context) {
				c.toastType = "success"
				c.toastText = toastText
				c.toastShow = (toastText != "")

				c.newFigFile = file.Get("name").String()
				c.newFigData = data

				c.user.AvatarURL = avatar
			})
			return
		}
	})
}
