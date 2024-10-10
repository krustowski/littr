package common

const (
	// generic error messages
	ERR_CALLER_BLANK     = "callerID cannot be empty"
	ERR_CALLER_FAIL      = "could not get caller's name"
	ERR_CALLER_NOT_FOUND = "caller not found in the database"
	ERR_USER_NOT_FOUND   = "user not found in the database"
	ERR_PAGENO_INCORRECT = "pageNo has to be specified as integer/number"
	ERR_PAGE_EXPORT_NIL  = "could not get more pages, one exported map is nil!"
	ERR_INPUT_DATA_FAIL  = "could not process the input data, try again"

	// auth-related error messages
	ERR_AUTH_FAIL           = "wrong credentials entered, or such user does not exist"
	ERR_AUTH_ACC_TOKEN_FAIL = "could not generate new access token"
	ERR_AUTH_REF_TOKEN_FAIL = "could not generate new refresh token"
	ERR_TOKEN_SAVE_FAIL     = "could not save new token to database"

	// post-related error messages
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

	// image-processing-related error messages
	ERR_IMG_DECODE_FAIL      = "image: could not decode to byte stream"
	ERR_IMG_ENCODE_FAIL      = "image: could not re-encode"
	ERR_IMG_ORIENTATION_FAIL = "image: could not fix the orientation"
	ERR_IMG_GIF_TO_WEBP_FAIL = "image: could not convert GIF to WebP"
	ERR_IMG_UNKNOWN_TYPE     = "image: unsupported format entered"
	ERR_IMG_SAVE_FILE_FAIL   = "image: could not save to a file"
	ERR_IMG_THUMBNAIL_FAIL   = "image: could not re-encode the thumbnail"

	// user-related error messages
	ERR_USER_DELETE_FOREIGN      = "you cannot delete a foreigner's account"
	ERR_USER_DELETE_FAIL         = "could not delete the user from user database, try again"
	ERR_SUBSCRIPTION_DELETE_FAIL = "could not delete the user from subscriptions, try again"

	// poll-related error messages
	ERR_POLL_AUTHOR_MISMATCH = "you cannot post a foreigner's poll"
	ERR_POLL_SAVE_FAIL       = "could not save the poll, try again"
	ERR_POLL_POST_FAIL       = "could not save a post about the new poll"
	ERR_POLL_NOT_FOUND       = "such poll not found in the database (may be deleted)"
	ERR_POLL_SELF_VOTE       = "you cannot vote on yours own poll"
	ERR_POLL_EXISTING_VOTE   = "you have already voted on such poll"
	ERR_POLL_DELETE_FOREIGN  = "you cannot delete a foreigner's poll"
	ERR_POLL_DELETE_FAIL     = "could not delete the poll, try again"

	// push-related error messages
	ERR_DEVICE_NOT_FOUND       = "devices not found in the database"
	ERR_SUBSCRIPTION_SAVE_FAIL = "could not save the subscription to database"
	ERR_SUBSCRIPTION_NOT_FOUND = "such subscription not found in the database"
	ERR_DEVICE_SUBSCRIBED      = "this device has been already registered"
	ERR_PUSH_SELF_NOTIF        = "will not notify oneself"
	ERR_PUSH_DISABLED_NOTIF    = "will not notify original poster"
	ERR_PUSH_UUID_BLANK        = "device's UUID cannot be sent empty"
)
