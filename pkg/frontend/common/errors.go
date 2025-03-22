// Common functions, constants, and structures package for the frontend.
package common

const (
	StateNameAuthGranted string = "auth-granted"
	StateNameNewUpdate   string = "new-update"
	StateNameUser        string = "user-data"
	StateNameCachedPosts string = "cached-posts"
)

const (
	// Event-related (non-)error messages.
	MSG_SERVER_START   = "The server has just (re)started."
	MSG_SERVER_RESTART = "The server is restarting now..."
	MSG_NEW_POLL       = "New poll has been just added."
	MSG_NEW_POST       = "New post added by %s."
	MSG_STATE_OFFLINE  = "You have gone offline. Check your Internet connection."
	MSG_STATE_ONLINE   = "You are back online."

	// Flow/Posts-related error messages.
	MSG_DELETE_SUCCESS      = "Post deleted."
	MSG_REPLY_ADDED         = "Reply added."
	MSG_EMPTY_FLOW          = "This flow is very empty, you can try expanding it."
	MSG_USER_HAS_NOT_POSTED = "This user has apparently not published any post yet."
	ERR_INVALID_REPLY       = "No valid content was entered."
	ERR_POST_UNAUTH_DELETE  = "You can only delete your own posts."
	ERR_POST_NOT_FOUND      = "Post not found (may be deleted)."
	ERR_USER_NOT_FOUND      = "User not found."
	ERR_PRIVATE_ACC         = "This account is private."
	ERR_INVALID_REQ_PARAMS  = "Invalid page request params."

	// Generic error messages on the FE.
	ERR_CANNOT_REACH_BE = "littr can't connect to the server."
	ERR_CANNOT_GET_DATA = "littr can't read the data."
	ERR_LOGIN_AGAIN     = "Please log in again."

	// Login-related error messages.
	MSG_USER_ACTIVATED          = "The user has been successfully activated, try logging in."
	ERR_ALL_FIELDS_REQUIRED     = "All fields are required."
	ERR_LOGIN_CHARS_LIMIT       = "Only characters a-z, A-Z and numbers can be used."
	ERR_ACCESS_DENIED           = "Incorrect login details entered."
	ErrLocalStorageUserSave     = "littr can't save user data to your browser."
	ErrLocalStorageUserLoad     = "littr can't save user data to your browser."
	ERR_ACTIVATION_INVALID_UUID = "Invalid activation UUID entered."
	ERR_ACTIVATION_EXPIRED_UUID = "Expired activation UUID entered."

	// Notification-related (non-)error messages.
	MSG_SUBSCRIPTION_UPDATED      = "Subscription updated."
	MSG_UNSUBSCRIBED_SUCCESS      = "Successfully unsubscribed, notifications are off."
	ERR_SUBSCRIPTION_UPDATE_FAIL  = "Your subscription could not be updated, please try again later."
	ERR_NOTIF_PERMISSION_DENIED   = "User notification denied."
	ERR_NOTIF_UNSUPPORTED_BROWSER = "Notifications are not supported in this browser."

	// Polls-related (non-)error messages.
	MSG_NO_POLL_TO_SHOW      = "No poll to display, click here to create one."
	ERR_POLL_UNAUTH_DELETE   = "You can delete your own polls only."
	ERR_POLL_OPTION_MISMATCH = "Such an option is not associated with this poll."

	// Post-related (non-)error messages.
	MSG_IMAGE_READY             = "Image is ready to upload."
	ERR_LOCAL_STORAGE_LOAD_FAIL = "Unable to decode user data."
	ERR_POST_TEXTAREA_EMPTY     = "No valid content was entered."
	ERR_POLL_FIELDS_REQUIRED    = "A poll question and at least two options are required."
	ERR_POST_UNKNOWN_TYPE       = "Unknown post type."

	// Register-related error messages.
	MSG_REGISTER_SUCCESS         = "Registration successful, check your mail inbox to get the activation link."
	ERR_REGISTER_FIELDS_REQUIRED = "All fields are required."
	ERR_REGISTER_CHARSET_LIMIT   = "Only characters a-z, A-Z and numbers can be used."
	ERR_WRONG_EMAIL_FORMAT       = "Invalid e-mail address format entered."

	// Reset-related (non-)error messages.
	MSG_RESET_PASSPHRASE_SUCCESS = "Your passphrase has been changed, check your e-mail inbox."
	MSG_RESET_REQUEST_SUCCESS    = "A request to reset your passphrase has been sent, please check your inbox."
	ERR_RESET_FIELD_REQUIRED     = "E-mail address is required."
	ERR_RESET_UUID_FIELD_EMPTY   = "An UUID string is required to continue, please check your inbox."
	ERR_RESET_INVALID_INPUT_DATA = "Invalid input data entered."

	// Settings-related (non-)error messages.
	MSG_PASSPHRASE_UPDATED       = "The passphrase was successfully updated."
	MSG_ABOUT_TEXT_UPDATED       = "The About text has been successfully updated."
	MSG_WEBSITE_UPDATED          = "The website has been successfully updated."
	MSG_SUBSCRIPTION_REQ_SUCCESS = "Successfully subscribed to notifications."
	MSG_UI_MODE_TOGGLE           = "UI mode changed."
	MSG_UI_THEME_TOGGLE          = "UI theme changed."
	MSG_LIVE_MODE_TOGGLE         = "Live mode changed."
	MSG_LOCAL_TIME_TOGGLE        = "Local time mode changed."
	MSG_PRIVATE_MODE_TOGGLE      = "Private mode changed."
	MSG_AVATAR_CHANGE_SUCCESS    = "Avatar has been successfully updated."
	ERR_PASSPHRASE_MISMATCH      = "Passphrases do not match."
	ERR_PASSPHRASE_MISSING       = "The passphrase field must be filled in."
	ERR_ABOUT_TEXT_UNCHANGED     = "The About text area is empty or the text has not changed."
	ERR_ABOUT_TEXT_CHAR_LIMIT    = "The About text must be less than 100 characters."
	ERR_WEBSITE_UNCHANGED        = "Website URL must be filled in or changed."
	ERR_WEBSITE_REGEXP_FAIL      = "Failed to check the website format."
	ERR_WEBSITE_INVALID          = "The website is probably not a valid URL."
	ERR_SUBSCRIPTION_BLANK_UUID  = "Blank UUID string."
	ERR_SUBSCRIPTION_REQ_FAIL    = "Failed to subscribe to notifications: "

	// Users-related (non-)error messages.
	MSG_USER_UPDATED_SUCCESS   = "User updated, request deleted."
	MSG_FOLLOW_REQUEST_REMOVED = "Follow request deleted."
	MSG_REQ_TO_FOLLOW_SUCCESS  = "The follow request was sent successfully."
	MSG_USER_FOLLOW_ADD_FMT    = "User %s followed now."
	MSG_USER_FOLLOW_REMOVE_FMT = "User %s unfollowed."
	MSG_SHADE_SUCCESSFUL       = "User was (un)shaded successfully."
)
