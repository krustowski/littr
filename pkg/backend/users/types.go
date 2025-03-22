package users

import (
	"go.vxn.dev/littr/pkg/models"
)

type UserActivationRequest struct {
	UUID string `json:"uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type UserCreateRequest struct {
	Email           string `json:"email" example:"alice@example.com"`
	Nickname        string `json:"nickname" example:"alice"`
	PassphrasePlain string `json:"passphrase_plain" example:"s3creTpassuWuort"`
}

type UserPassphraseRequest struct {
	// Passphrase reset pre-request
	Email string `json:"email" example:"alice@example.com"`
}

type UserPassphraseReset struct {
	// Passphrase reset request
	UUID string `json:"uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type UserUpdateListsRequest struct {
	// Lists update request payload.
	FlowList    map[string]bool `json:"flow_list" example:"bob:false"`
	RequestList map[string]bool `json:"request_list" example:"cody:true"`
	ShadeList   map[string]bool `json:"shade_list" example:"dave:true"`
}

type UserUpdateOptionsRequest struct {
	// Options updata request payload (legacy fields).
	UIMode        bool                  `json:"ui_mode"`
	UITheme       models.Theme          `json:"ui_theme"`
	LiveMode      bool                  `json:"live_mode"`
	LocalTimeMode bool                  `json:"local_time_mode"`
	Private       bool                  `json:"private"`
	AboutText     string                `json:"about_you" example:"let's gooo"`
	WebsiteLink   string                `json:"website_link" example:"https://example.com"`
	OptionsMap    models.UserOptionsMap `json:"options_map" example:"private:true"`
}

type UserUpdatePassphraseRequest struct {
	// New passphrase request payload.
	NewPassphrase     string `json:"new_passphrase_plain"`
	CurrentPassphrase string `json:"current_passphrase_plain"`
}

type UserUpdateSubscriptionRequest []string

type UserUploadAvatarRequest struct {
	// New avatar upload/update request payload.
	AvatarByteData []byte `json:"data" format:"base64"`
	AvatarFileName string `json:"figure" example:"avatar.jpeg"`
}
