package common

const (
	// toast types
	TTYPE_ERR = "error"
	TTYPE_INFO = "info"
	TTYPE_SUCCESS = "success"
)

const (
	// generic error messages on FE
	ERR_CANNOT_REACH_BE = "cannot reach the server"
	ERR_CANNOT_GET_DATA = "cannot get the data"
	ERR_LOGIN_AGAIN = "please log-in again"

	// flow/post-related error messages
	ERR_INVALID_REPLY = "no valid reply content entered"
	ERR_POST_UNAUTH_DELETE = "you can delete your own posts only!"
	ERR_POST_NOT_FOUND = "post not found"
	ERR_USER_NOT_FOUND = "user not found"
	ERR_PRIVATE_ACC = "this account is private"

	// notification-related (non-)error messages
	MSG_SUBSCRIPTION_UPDATED = "subscription updated"
	MSG_UNSUBSCRIPED_SUCCESS = "successfully unsubscribed, notifications off"
	ERR_SUBSCRIPTION_UPDATE_FAIL = "failed to update the subscription, try again later"
	ERR_NOTIF_PERMISSION_DENIED = "notification permission denied by user"
	ERR_NOTIF_UNSUPPORTED_BROWSER = "notifications are not supported in this browser"

	// settings-related (non-)error messages
	MSG_PASSPHRASE_UPDATED = "passphrase updated successfully"
	MSG_ABOUT_TEXT_UPDATED = "about text updated successfully"
	MSG_WEBSITE_UPDATED = "website updated successfully"
	MSG_UNSUBSCRIBED_SUCCESS = "device unsubscribed successfully"
	MSG_SUBSCRIPTION_REG_SUCCESS = "successfully subscribed to notifs"
	MSG_LOCAL_TIME_TOGGLE = "local time mode toggled"
	MSG_PRIVATE_MODE_TOGGLE = "private mode toggled"
	MSG_AVATAR_CHANGE_SUCCESS = "avatar updated successfully"
	ERR_PASSPHRASE_MISMATCH = "passphrases do not match"
	ERR_PASSPHRASE_MISSING = "passphrase fields need to be filled"
	ERR_ABOUT_TEXT_UNCHANGED = "the about textarea is empty, or the text has not changed"
	ERR_ABOUT_TEXT_CHAR_LIMIT = "about text has to be shorter than 100 chars"
	ERR_WEBSITE_UNCHANGED = "website URL has to be filled, or changed"
	ERR_WEBSITE_REGEXP_FAIL = "failed to check the website format (regexp object failed)"
	ERR_WEBSITE_INVALID = "website prolly not a valid URL"
	ERR_SUBSCRIPTION_BLANK_UUID = "blank UUID string"
	ERR_SUBSCRIPTION_REG_FAIL = "failed to subscribe to notifications: "

	// login-related error messages
	ERR_ALL_FIELDS_REQUIRED = "all fields are required"
	ERR_LOGIN_CHARS_LIMIT = "only a-z, A-Z characters and numbers can be used"
	ERR_ACCESS_DENIED = "wrong credentials entered"
	ERR_LOCAL_STORAGE_USER_FAIL = "cannot save user's data to the browser"

	// polls-related (non-)error messages
	MSG_NO_POLL_TO_SHOW = "no poll to show, click here to create one!"
	ERR_POLL_UNAUTH_DELETE = "you can delete your own polls only!"
	ERR_POLL_OPTION_MISMATCH = "such option is not associated to the poll"
)
