package push

import (
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/helpers"
	"go.vxn.dev/littr/pkg/models"

	"github.com/SherClockHolmes/webpush-go"
)

type NotificationOpts struct {
	Receiver string
	Devices  *[]models.Device
	Body     *[]byte
	Repo     models.UserRepositoryInterface
}

func SendNotificationToDevices(opts *NotificationOpts) {
	l := common.NewLogger(nil, "notifs")
	//l := opts.Logger
	stringifiedBody := string(*opts.Body)

	var tag string

	if strings.Contains(stringifiedBody, "reply") {
		tag = "reply"
	} else if strings.Contains(stringifiedBody, "mention") {
		tag = "mention"
	}

	// prepare an array for possible invalid devices (expired subscriptions etc)
	devicesToDelete := []string{}

	var wg sync.WaitGroup

	// range devices and fire notifications
	for _, dev := range *opts.Devices {
		// skip blank UUIDs
		if dev.UUID == "" {
			continue
		}

		// skip devices unsubscribed to such notification tag
		if !helpers.Contains(dev.Tags, tag) {
			continue
		}

		wg.Add(1)

		// run this async not to make client wait too much
		//
		// IMPORTANT NOTE: DO NOT write headers in the goroutine --- this will crash the server on nil pointer dereference
		// and memory segment violation
		go func(dev models.Device, body *[]byte) {
			defer wg.Done()

			sub := dev.Subscription

			websub := &webpush.Subscription{
				Endpoint: sub.Endpoint,
				Keys: webpush.Keys{
					Auth:   sub.Keys.Auth,
					P256dh: sub.Keys.P256dh,
				},
			}
			// fire the notification
			res, err := webpush.SendNotification(*body, websub, &webpush.Options{
				Subscriber:      os.Getenv("VAPID_SUBSCRIBER"),
				VAPIDPublicKey:  os.Getenv("VAPID_PUB_KEY"),
				VAPIDPrivateKey: os.Getenv("VAPID_PRIV_KEY"),
				TTL:             30,
				Urgency:         webpush.UrgencyNormal,
			})
			if err != nil {
				l.Msg(common.ERR_NOTIFICATION_NOT_SENT + err.Error()).Status(http.StatusInternalServerError).Log()
				return
			}

			defer func() {
				if err := res.Body.Close(); err != nil {
					l.Error(err).Log()
				}
			}()

			// read the response body
			bodyBytes, err := io.ReadAll(res.Body)
			if err != nil {
				l.Msg(common.ERR_NOTIFICATION_RESP_BODY_FAIL + err.Error()).Status(http.StatusInternalServerError).Log()
				return
			}

			// stringify the text response
			bodyString := string(bodyBytes)
			if bodyString == "" {
				bodyString = "okay"
			}

			// successful notification processing (webpush) gateway's response is HTTP/201
			// otherwise is expired or unsubscribed => delete subscription
			if res.StatusCode != 201 {
				devicesToDelete = append(devicesToDelete, dev.UUID)
			}

			l.Msg(common.MSG_WEBPUSH_GW_RESPONSE + bodyString).Status(res.StatusCode).Log()
		}(dev, opts.Body)
	}

	wg.Wait()

	// update device list --- do not include devs to delete (ones failed to send the notification to)
	defer func(devs *[]models.Device, oldUUIDs []string) {
		// no invalid devices = no worries
		if len(devicesToDelete) == 0 {
			return
		}

		// prepare a new device array
		newDeviceList := []models.Device{}

		// loop over devs to cherrypick the currently valid devs
		for _, dev := range *devs {
			if helpers.Contains(oldUUIDs, dev.UUID) {
				continue
			}

			dev.TimeLastUsed = time.Now()
			newDeviceList = append(newDeviceList, dev)
		}

		receiver, err := opts.Repo.GetByID(opts.Receiver)
		if err != nil {
			l.Msg(common.ERR_DEVICE_LIST_UPDATE_FAIL).Error(err).Status(http.StatusInternalServerError).Log()
			return
		}

		receiver.Devices = newDeviceList

		// save new device array upon the callerID
		if err := opts.Repo.Save(receiver); err != nil {
			l.Msg(common.ERR_DEVICE_LIST_UPDATE_FAIL).Error(err).Status(http.StatusInternalServerError).Log()
			return
		}

		l.Msg("ok, device list updated").Status(http.StatusOK).Log()
	}(opts.Devices, devicesToDelete)
}
