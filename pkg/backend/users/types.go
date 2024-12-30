package users

import (
	"go.vxn.dev/littr/pkg/models"
)

type UserCreateRequest struct {
	Nickname      string `json:"nickname" example:"alice"`
	PassphraseHex string `json:"passphrase_hex" example:"fb43b35a752b0e8045e2dd1b1e292983b9cbf4672a51e30caaa3f9b06c5a3b74d5096bc8092c9e90a2e047c1eab29eceb50c09d6c51e6995c1674beb3b06535e"`
	Email         string `json:"email" example:"alice@example.com"`
}

type UserPassphraseResetRequest struct {
	// Passphrase reset request
	UUID  string `json:"uuid"`
	Email string `json:"email" example:"alice@example.com"`
}

type UserUpdateListsRequest struct {
	// Lists update request payload.
	FlowList    map[string]bool `json:"flow_list" example:"bob:false"`
	RequestList map[string]bool `json:"request_list" example:"cody:true"`
	ShadeList   map[string]bool `json:"shade_list" example:"dave:true"`
}

type UserUpdateOptionsRequest struct {
	// Options updata request payload (legacy fields).
	UIDarkMode    bool                  `json:"dark_mode"`
	LiveMode      bool                  `json:"live_mode"`
	LocalTimeMode bool                  `json:"local_time_mode"`
	Private       bool                  `json:"private"`
	AboutText     string                `json:"about_you" example:"let's gooo"`
	WebsiteLink   string                `json:"website_link" example:"https://example.com"`
	OptionsMap    models.UserOptionsMap `json:"options_map" example:"private:true"`
}

type UserUpdatePassphraseRequest struct {
	// New passphrase request payload.
	NewPassphraseHex     string `json:"new_passphrase_hex"`
	CurrentPassphraseHex string `json:"current_passphrase_hex"`
}

type UserUploadAvatarRequest struct {
	// New avatar upload/update request payload.
	AvatarByteData []byte `json:"data" format:"base64"`
	AvatarFileName string `json:"figure" example:"avatar.jpeg"`
}
