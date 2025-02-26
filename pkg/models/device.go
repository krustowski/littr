// The very models-related package containing all the in-database-saved types and structures.
package models

import (
	"time"
)

type Devices []Device

// SubscriptionDevice
type Device struct {
	// Unique identification of the app on the current device.
	// https://go-app.dev/reference#Context
	UUID string `json:"uuid"`

	// Nickname is the identificator of such device's possessor.
	Nickname string `json:"nickname"`

	// Timestamp of the subscription creation.
	TimeCreated time.Time `json:"time_created"`

	// Timestamp of the last notification sent through this device.
	TimeLastUsed time.Time `json:"time_last_used"`

	// List of labels for such device.
	Tags []string `json:"tags,omitempty"`

	// The very subscription struct/details.
	Subscription Subscription `json:"subscription"`
}

type Subscription struct {
	Endpoint string `json:"endpoint"`
	Keys     Keys   `json:"keys"`
}

type Keys struct {
	Auth   string `json:"auth"`
	P256dh string `json:"p256dh"`
}

func (dd Devices) GetID() string {
	for _, dev := range dd {
		if dev.Nickname != "" {
			return dev.Nickname
		}
	}
	return ""
}
