package settings

import (
	"crypto/sha512"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/models"

	//"github.com/SherClockHolmes/webpush-go"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func (c *Content) onAvatarChange(ctx app.Context, e app.Event) {
	ctx.NewActionWithValue("avatar-change", e.Get("target").Get("files").Index(0))
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
			toast.Text(common.ERR_PASSPHRASE_MISSING).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		if passphrase != passphraseAgain {
			toast.Text(common.ERR_PASSPHRASE_MISMATCH).Type(common.TTYPE_ERR).Dispatch()
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
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		c.user.Passphrase = string(passHash[:])

		/*var userStream []byte
		if err := reload(c.user, &userStream); err != nil {
			toast.Text("cannot update local storage").Type("error").Dispatch()

			ctx.Dispatch(func(ctx app.Context) {
				c.settingsButtonDisabled = false
			})
			return
		}*/

		toast.Text(common.MSG_PASSPHRASE_UPDATED).Type(common.TTYPE_SUCCESS).Dispatch()
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
			toast.Text(common.ERR_ABOUT_TEXT_UNCHANGED).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		if len(aboutText) > 100 {
			toast.Text(common.ERR_ABOUT_TEXT_CHAR_LIMIT).Type(common.TTYPE_ERR).Dispatch()
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
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		// Update the LocalStorage.
		common.SaveUser(&c.user, &ctx)

		toast.Text(common.MSG_ABOUT_TEXT_UPDATED).Type(common.TTYPE_SUCCESS).Dispatch()
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
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		toast.Text(common.MSG_WEBSITE_UPDATED).Type(common.TTYPE_SUCCESS).Dispatch()

		// Update the LocalStorage.
		common.SaveUser(&c.user, &ctx)

		ctx.Dispatch(func(ctx app.Context) {
			// update user's struct in memory
			c.user.Web = c.website
		})
	})
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
		toast.Text(common.ERR_SUBSCRIPTION_BLANK_UUID).Type(common.TTYPE_ERR).Dispatch()
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
			Url:         "/api/v1/push/subscriptions/" + ctx.DeviceID(),
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

// onClickDeleteAccount()
