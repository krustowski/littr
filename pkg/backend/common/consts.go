// Common functions, constants, and structures package for the backend.
package common

import (
	"time"
)

// Token consts.
const (
	// Refresh token's TTL.
	TOKEN_TTL = time.Hour * 168 * 4
)

// Header names.
const (
	HDR_PAGE_NO      = "X-Page-No"
	HDR_HIDE_REPLIES = "X-Hide-Replies"
	HDR_DUMP_TOKEN   = "X-Dump-Token"
	DHR_API_TOKEN    = "X-API-Token"
)

// (Non-)Error messages.
const (
	// Auth-related error messages
	ERR_AUTH_FAIL           = "wrong credentials entered, or such user does not exist"
	ERR_AUTH_ACC_TOKEN_FAIL = "could not generate new access token"
	ERR_AUTH_REF_TOKEN_FAIL = "could not generate new refresh token"
	ERR_TOKEN_SAVE_FAIL     = "could not save new token to database"

	// Generic error messages
	ERR_CALLER_BLANK      = "callerID cannot be empty"
	ERR_CALLER_FAIL       = "could not get caller's name"
	ERR_CALLER_NOT_FOUND  = "caller not found in the database"
	ERR_USER_NOT_FOUND    = "user not found in the database"
	ERR_PAGENO_INCORRECT  = "pageNo has to be specified as integer/number"
	ERR_PAGE_EXPORT_NIL   = "could not get more pages, one exported map is nil!"
	ERR_INPUT_DATA_FAIL   = "could not process the input data, try again"
	ERR_API_TOKEN_BLANK   = "blank API token sent"
	ERR_API_TOKEN_INVALID = "invalid API token sent"
	ERR_NO_SERVER_SECRET  = "missing the server's secret (APP_PEPPER)"

	// Image-processing-related error messages
	ERR_IMG_DECODE_FAIL      = "image: could not decode to byte stream"
	ERR_IMG_ENCODE_FAIL      = "image: could not re-encode"
	ERR_IMG_ORIENTATION_FAIL = "image: could not fix the orientation"
	ERR_IMG_GIF_TO_WEBP_FAIL = "image: could not convert GIF to WebP"
	ERR_IMG_UNKNOWN_TYPE     = "image: unsupported format entered"
	ERR_IMG_SAVE_FILE_FAIL   = "image: could not save to a file"
	ERR_IMG_THUMBNAIL_FAIL   = "image: could not re-encode the thumbnail"

	// Poll-related error messages
	ERR_POLL_AUTHOR_MISMATCH    = "you cannot post a foreigner's poll"
	ERR_POLL_SAVE_FAIL          = "could not save the poll, try again"
	ERR_POLL_POST_FAIL          = "could not save a post about the new poll"
	ERR_POLL_NOT_FOUND          = "such poll not found in the database (may be deleted)"
	ERR_POLL_SELF_VOTE          = "you cannot vote in yours own poll"
	ERR_POLL_EXISTING_VOTE      = "you have already voted on such poll"
	ERR_POLL_DELETE_FOREIGN     = "you cannot delete a foreigner's poll"
	ERR_POLL_DELETE_FAIL        = "could not delete the poll, try again"
	ERR_POLLID_BLANK            = "pollID param is required"
	ERR_POLL_INVALID_VOTE_COUNT = "you can pass only one vote per poll"

	// Post-related error messages
	ERR_POST_BLANK          = "post has got no content"
	ERR_POSTER_INVALID      = "you can add yours posts only"
	ERR_POST_SAVE_FAIL      = "could not save the post, try again"
	ERR_POST_NOT_FOUND      = "could not find the post (may be deleted)"
	ERR_POST_SELF_RATE      = "you cannot rate your own posts"
	ERR_POST_UPDATE_FOREIGN = "you cannot update a foreigner's post"
	ERR_POST_DELETE_FOREIGN = "you cannot delete a foreigner's post"
	ERR_POST_DELETE_FAIL    = "could not delete the post, try again"
	ERR_POST_DELETE_THUMB   = "could not delete associated thumbnail"
	ERR_POST_DELETE_FULLIMG = "could not delete associated full image"
	ERR_POSTID_BLANK        = "postID param is required"
	ERR_HASHTAG_BLANK       = "hashtag param is required"

	// Push-related (non-)error messages
	MSG_WEBPUSH_GW_RESPONSE         = "push goroutine: webpush gateway:"
	ERR_DEVICE_NOT_FOUND            = "devices not found in the database"
	ERR_SUBSCRIPTION_SAVE_FAIL      = "could not save the subscription to database"
	ERR_SUBSCRIPTION_NOT_FOUND      = "such subscription not found in the database"
	ERR_DEVICE_SUBSCRIBED_ALREADY   = "this device has been already registered"
	ERR_PUSH_SELF_NOTIF             = "will not notify oneself"
	ERR_PUSH_DISABLED_NOTIF         = "will not notify original poster"
	ERR_PUSH_UUID_BLANK             = "device's UUID cannot be sent empty"
	ERR_PUSH_BODY_COMPOSE_FAIL      = "failed to compose the notification body"
	ERR_NOTIFICATION_NOT_SENT       = "notification could not be sent"
	ERR_NOTIFICATION_RESP_BODY_FAIL = "failed to read the notification's response body"
	ERR_DEVICE_LIST_UPDATE_FAIL     = "failed to save the updated device list"
	ERR_UUID_BLANK                  = "uuid param is required"

	// User-related error messages
	ERR_USER_DELETE_FOREIGN       = "you cannot delete a foreigner's account"
	ERR_USER_DELETE_FAIL          = "could not delete the user from user database, try again"
	ERR_SUBSCRIPTION_DELETE_FAIL  = "could not delete the user from subscriptions, try again"
	ERR_MISSING_IMG_CONTENT       = "no image data received, try again"
	ERR_REGISTRATION_DISABLED     = "registration is disabled at the moment"
	ERR_RESTRICTED_NICKNAME       = "this nickname is restricted"
	ERR_USER_NICKNAME_TAKEN       = "this nickname has been already taken"
	ERR_NICKNAME_CHARSET_MISMATCH = "the nickname can only consist of characters a-z, A-Z and numbers 0-9"
	ERR_NICKNAME_TOO_LONG_SHORT   = "nickname is too long (>12 runes), or too short (<3 runes)"
	ERR_WRONG_EMAIL_FORMAT        = "wrong format of the e-mail address"
	ERR_EMAIL_ALREADY_USED        = "this e-mail address has been already used"
	ERR_USER_SAVE_FAIL            = "could not save new user to database"
	ERR_POSTREG_POST_SAVE_FAIL    = "could not save a new post about the new user's addition"
	ERR_USER_UPDATE_FOREIGN       = "could not update a foreign's account"
	ERR_USERID_BLANK              = "userID param is required"
	ERR_USER_UPDATE_FAIL          = "could not update the user in database"
	ERR_USER_PASSPHRASE_FOREIGN   = "you can change yours passphrase only"
	ERR_PASSPHRASE_REQ_INCOMPLETE = "passphrase change request is partly or completely empty"
	ERR_PASSPHRASE_CURRENT_WRONG  = "current passphrase sent is wrong"
	ERR_NO_EMAIL_MATCH            = "could not find a match for such e-mail address"
	ERR_MAIL_COMPOSITION_FAIL     = "could not compose the mail properly, try again"
	ERR_MAIL_NOT_SENT             = "could not send the mail, try again"
	ERR_REQUEST_UUID_BLANK        = "could not process blank request's UUID"
	ERR_REQUEST_EMAIL_BLANK       = "could not process blank request's e-mail address"
	ERR_REQUEST_UUID_INVALID      = "entered UUID is invalid"
	ERR_REQUEST_UUID_EXPIRED      = "entered UUID has expired"
	ERR_REQUEST_DELETE_FAIL       = "could not delete the request from database, try again"
	ERR_PASSPHRASE_UPDATE_FAIL    = "could not update the passphrase in database, try again"
	ERR_TARGET_USER_NOT_PRIVATE   = "target user is not private, no need to file new follow requests"
	ERR_USER_AVATAR_FOREIGN       = "you can update yours avatar only"
	ERR_ACTIVATION_MAIL_FAIL      = "the activation mail was not sent, try again"
	ERR_USER_NOT_ACTIVATED        = "user has not been activated yet, check your mail inbox"

)
