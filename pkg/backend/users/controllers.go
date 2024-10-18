package users

import (
	"crypto/sha512"
	"fmt"
	"math/rand"
	"net/http"
	netmail "net/mail"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.vxn.dev/littr/configs"
	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/backend/db"
	"go.vxn.dev/littr/pkg/backend/image"
	"go.vxn.dev/littr/pkg/backend/mail"
	"go.vxn.dev/littr/pkg/backend/pages"
	"go.vxn.dev/littr/pkg/helpers"
	"go.vxn.dev/littr/pkg/models"

	chi "github.com/go-chi/chi/v5"
	uuid "github.com/google/uuid"
)

// getUsers is the users handler that processes and returns existing users list.
//
// @Summary      Get a list of users
// @Description  get a list of users
// @Tags         users
// @Produce      json
// @Param    	 X-Page-No header string true "page number"
// @Success      200  {object}   common.APIResponse{data=users.getUsers.responseData}
// @Failure	 400  {object}   common.APIResponse
// @Failure	 404  {object}   common.APIResponse
// @Failure	 500  {object}   common.APIResponse
// @Router       /users [get]
func getUsers(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "users")

	type responseData struct {
		Users     map[string]models.User     `json:"users"`
		User      models.User                `json:"user"`
		UserStats map[string]models.UserStat `json:"user_stats"`
	}

	// skip blank callerID
	if l.CallerID == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// check the callerID record in the database
	if _, ok := db.GetOne(db.UserCache, l.CallerID, models.User{}); !ok {
		l.Msg(common.ERR_CALLER_NOT_FOUND).Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	// fetch the required X-Page-No header
	pageNoString := r.Header.Get(common.HDR_PAGE_NO)
	pageNo, err := strconv.Atoi(pageNoString)
	if err != nil {
		l.Msg(common.ERR_PAGENO_INCORRECT).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	opts := pages.PageOptions{
		CallerID: l.CallerID,
		PageNo:   pageNo,
		FlowList: nil,

		Users: pages.UserOptions{
			Plain: true,
		},
	}

	// fetch one (1) users page
	pagePtrs := pages.GetOnePage(opts)
	if pagePtrs == (pages.PagePointers{}) || pagePtrs.Users == nil || (*pagePtrs.Users) == nil {
		l.Msg(common.ERR_PAGE_EXPORT_NIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	//(*pagePtrs.Users)[l.CallerID] = caller

	// fetch all data for the calculations
	posts, _ := db.GetAll(db.FlowCache, models.Post{})

	// prepare a map for statistics
	stats := make(map[string]models.UserStat)

	// calculate the post count
	for _, post := range posts {
		// is there already a statistic struct for such user?
		stat, found := stats[post.Nickname]
		if !found {
			stat = models.UserStat{}
		}

		// increase the post count and assign the stat back to stats map
		stat.PostCount++
		stats[post.Nickname] = stat
	}

	// calculate the flower/follower count
	for nick, user := range *pagePtrs.Users {
		// get one's flowList
		flowList := user.FlowList
		if flowList == nil {
			continue
		}

		// loop over all flowList items, increment those followed
		for key, state := range flowList {
			if state && key != nick {
				// increment the follower count
				stat := stats[key]
				stat.FlowerCount++
				stats[key] = stat
			}
		}
	}

	// flush unwanted properties in users map
	users := *common.FlushUserData(pagePtrs.Users, l.CallerID)

	// prepare the response payload
	pl := &responseData{
		Users:     users,
		User:      users[l.CallerID],
		UserStats: stats,
	}

	l.Msg("ok, dumping users").Status(http.StatusOK).Log().Payload(pl).Write(w)
	return
}

// getOneUser is the users handler that processes and returns existing user's details according to callerID.
//
// @Summary      Get the user's details
// @Description  get the user's details
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}   common.APIResponse{data=users.getOneUser.responseData}
// @Failure      400  {object}   common.APIResponse
// @Failure      404  {object}   common.APIResponse
// @Router       /users/caller [get]
func getOneUser(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "users")

	type responseData struct {
		User      models.User     `json:"user"`
		Devices   []models.Device `json:"devices"`
		PublicKey string          `json:"public_key"`
	}

	// skip blank callerID
	if l.CallerID == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// check the callerID record in the database
	caller, ok := db.GetOne(db.UserCache, l.CallerID, models.User{})
	if !ok {
		l.Msg(common.ERR_USER_NOT_FOUND).Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	// fetch user's devices
	devs, _ := db.GetOne(db.SubscriptionCache, l.CallerID, []models.Device{})

	// helper struct for the data flush
	users := make(map[string]models.User)
	users[l.CallerID] = caller

	// flush sensitive user data
	users = *common.FlushUserData(&users, l.CallerID)

	// compose the response payloaad
	pl := &responseData{
		User:      users[l.CallerID],
		Devices:   devs,
		PublicKey: os.Getenv("VAPID_PUB_KEY"),
	}

	l.Status(http.StatusOK).Msg("ok, returning callerID's user record").Log().Payload(pl).Write(w)
	return
}

// addNewUser is the users handler that processes input and creates a new user.
//
// @Summary      Add new user
// @Description  add new user
// @Tags         users
// @Accept       json
// @Produce      json
// @Param    	 request body models.User true "new user's request body"
// @Success      200  {object}   common.APIResponse
// @Failure      400  {object}   common.APIResponse
// @Failure      403  {object}   common.APIResponse
// @Failure      409  {object}   common.APIResponse
// @Failure      500  {object}   common.APIResponse
// @Router       /users [post]
func addNewUser(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "users")

	// check if the registration is allowed
	if !configs.REGISTRATION_ENABLED {
		l.Msg(common.ERR_REGISTRATION_DISABLED).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	var user models.User

	// decode the incoming data
	if err := common.UnmarshalRequestData(r, &user); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// block restricted nicknames, use lowercase for comparison
	if helpers.Contains(configs.UserDeletionList, strings.ToLower(user.Nickname)) {
		l.Msg(common.ERR_RESTRICTED_NICKNAME).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	if _, found := db.GetOne(db.UserCache, user.Nickname, models.User{}); found {
		l.Msg(common.ERR_USER_NICKNAME_TAKEN).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// restrict the nickname to contains only some characters explicitly "listed"
	// https://stackoverflow.com/a/38554480
	if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(user.Nickname) {
		l.Msg(common.ERR_NICKNAME_CHARSET_MISMATCH).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// check the nick's length limits
	if len(user.Nickname) > 12 || len(user.Nickname) < 3 {
		l.Msg(common.ERR_NICKNAME_TOO_LONG_SHORT).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// get the e-mail address
	email := strings.ToLower(user.Email)
	user.Email = email

	// validate e-mail struct
	// https://stackoverflow.com/a/66624104
	if _, err := netmail.ParseAddress(email); err != nil {
		l.Msg(common.ERR_WRONG_EMAIL_FORMAT).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// check for the already registred e-mail
	allUsers, _ := db.GetAll(db.UserCache, models.User{})

	for _, usr := range allUsers {
		// e-maill address match
		if strings.ToLower(usr.Email) == user.Email {
			l.Msg(common.ERR_EMAIL_ALREADY_USED).Status(http.StatusConflict).Log().Payload(nil).Write(w)
			return
		}
	}

	//
	// validation end = new user can be added
	//

	// set defaults and a timestamp
	user.LastActiveTime = time.Now()
	user.About = "newbie"

	// options defaults
	user.Options = make(map[string]bool)
	user.Options["gdpr"] = true
	user.GDPR = true

	// save new user to the database
	if saved := db.SetOne(db.UserCache, user.Nickname, user); !saved {
		l.Msg(common.ERR_USER_SAVE_FAIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	// prepare a timestamp for a very new post to alert others the new user has been added
	postStamp := time.Now()
	postKey := strconv.FormatInt(postStamp.UnixNano(), 10)

	// compose the post's body
	post := models.Post{
		ID:        postKey,
		Type:      "user",
		Figure:    user.Nickname,
		Nickname:  "system",
		Content:   "new user has been added (" + user.Nickname + ")",
		Timestamp: postStamp,
	}

	// save new post to the database
	if saved := db.SetOne(db.FlowCache, postKey, post); !saved {
		l.Msg(common.ERR_POSTREG_POST_SAVE_FAIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("ok, adding new user").Status(http.StatusCreated).Log().Payload(nil).Write(w)
	return
}

// updateUser is the users handler function that processes and updates given user in the database.
//
// @Deprecated
// @Summary      Update the user's details
// @Description  update the user's details
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        userID path string true "updated user's ID"
// @Param    	 request body models.User true "updated user's body"
// @Success      200  {object}   common.APIResponse
// @Failure      400  {object}   common.APIResponse
// @Failure      404  {object}   common.APIResponse
// @Failure      409  {object}   common.APIResponse
// @Router       /users/{userID} [put]
func updateUser(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "users")

	// skip blank callerID
	if l.CallerID == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// take the param from path
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		l.Msg(common.ERR_USERID_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	var user models.User

	// decode the incoming data
	if err := common.UnmarshalRequestData(r, &user); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// mismatching user's name from the caller's ID
	if user.Nickname != l.CallerID || userID != l.CallerID || userID != user.Nickname {
		l.Msg(common.ERR_USER_UPDATE_FOREIGN).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	// get the uploaded user's database record verification
	if _, found := db.GetOne(db.UserCache, user.Nickname, models.User{}); !found {
		l.Msg(common.ERR_USER_NOT_FOUND).Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	// save uploaded user object as a whole --- dangerous and nasty!!! DO NOT use this handler
	if saved := db.SetOne(db.UserCache, user.Nickname, user); !saved {
		l.Msg(common.ERR_USER_UPDATE_FAIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("ok, user updated").Status(http.StatusOK).Log().Payload(nil).Write(w)
	return
}

// updateUserPassphrase is the users handler that allows the user to change their passphrase.
//
// @Summary      Update user's passphrase
// @Description  update user's passphrase
// @Tags         users
// @Produce      json
// @Param    	 request body users.updateUserPassphrase.requestData true "new passphrase hash body"
// @Param        userID path string true "ID of the user to update"
// @Success      200  {object}   common.APIResponse
// @Failure      400  {object}   common.APIResponse
// @Failure      403  {object}   common.APIResponse
// @Failure      404  {object}   common.APIResponse
// @Failure      409  {object}   common.APIResponse
// @Failure      500  {object}   common.APIResponse
// @Router       /users/{userID}/passphrase [patch]
func updateUserPassphrase(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "users")

	type requestData struct {
		NewPassphraseHex     string `json:"new_passphrase_hex"`
		CurrentPassphraseHex string `json:"current_passphrase_hex"`
	}

	// skip blank callerID
	if l.CallerID == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// take the param from path
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		l.Msg(common.ERR_USERID_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// check for possible user's record forgery attempt
	if l.CallerID != userID {
		l.Msg(common.ERR_USER_PASSPHRASE_FOREIGN).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	// fetch requested user's database record
	user, found := db.GetOne(db.UserCache, l.CallerID, models.User{})
	if !found {
		l.Msg(common.ERR_USER_NOT_FOUND).Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	var reqData requestData

	// decode the incoming data
	if err := common.UnmarshalRequestData(r, &reqData); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// check if both new or old passphrase hashes are blank/empty
	if reqData.NewPassphraseHex == "" || reqData.CurrentPassphraseHex == "" {
		l.Msg(common.ERR_PASSPHRASE_REQ_INCOMPLETE).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// check if the current passphrasë́'s hash is correct
	if reqData.CurrentPassphraseHex != user.PassphraseHex {
		l.Msg(common.ERR_PASSPHRASE_CURRENT_WRONG).Status(http.StatusConflict).Log().Payload(nil).Write(w)
		return
	}

	//
	// ok, passhprase should be okay to update/change
	//

	// set new passphrase hash to the user's record
	user.PassphraseHex = reqData.NewPassphraseHex

	// save the user's record back to database
	if saved := db.SetOne(db.UserCache, user.Nickname, user); !saved {
		l.Msg(common.ERR_USER_UPDATE_FAIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("ok, passphrase updated").Status(http.StatusOK).Log().Payload(nil).Write(w)
	return
}

// updateUserList is the users handler that allows one to update various lists associated with such one.
//
// @Summary      Update user's list
// @Description  update user's list
// @Tags         users
// @Produce      json
// @Param    	 request body users.updateUserList.requestData true "new user lists data"
// @Param        userID path string true "ID of the user to update"
// @Success      200  {object}   common.APIResponse
// @Failure      403  {object}   common.APIResponse
// @Failure      404  {object}   common.APIResponse
// @Failure      500  {object}   common.APIResponse
// @Router       /users/{userID}/lists [patch]
func updateUserList(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "users")

	type requestData struct {
		FlowList    map[string]bool `json:"flow_list"`
		RequestList map[string]bool `json:"request_list"`
		ShadeList   map[string]bool `json:"shade_list"`
	}

	// skip blank callerID
	if l.CallerID == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// take the param from path
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		l.Msg(common.ERR_USERID_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// check for possible user's record forgery attempt
	if l.CallerID != userID {
		l.Msg(common.ERR_USER_PASSPHRASE_FOREIGN).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	// look for the requested user's record in database
	user, found := db.GetOne(db.UserCache, userID, models.User{})
	if !found {
		l.Msg(common.ERR_USER_NOT_FOUND).Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	var reqData requestData

	// decode the incoming data
	if err := common.UnmarshalRequestData(r, &reqData); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	// process the FlowList if not empty
	if reqData.FlowList != nil {
		if user.FlowList == nil {
			user.FlowList = make(map[string]bool)
		}

		// loop over all flowList records
		for key, value := range reqData.FlowList {
			// do not allow to unfollow oneself
			if key == user.Nickname {
				user.FlowList[key] = true
				continue
			}

			// set such flowList record according to the request data
			if _, found := user.FlowList[key]; found {
				user.FlowList[key] = value
			}

			// check if the user is shaded by the counterpart
			if counterpart, exists := db.GetOne(db.UserCache, key, models.User{}); exists {
				if counterpart.Private && key != l.CallerID {
					// cannot add this user to one's flow, as the following
					// has to be requested and allowed by the counterpart
					user.FlowList[key] = false
					continue
				}

				// update the flowList record according to the counterpart's shade list state of the user
				if shaded, found := counterpart.ShadeList[user.Nickname]; !found || (found && !shaded) {
					user.FlowList[key] = value
				}
			}
		}
	}
	// always allow to see system posts
	user.FlowList["system"] = true

	// process the RequestList if not empty
	if reqData.RequestList != nil {
		if user.RequestList == nil {
			user.RequestList = make(map[string]bool)
		}

		// loop over the RequestList records and change the user's values accordingly
		for key, value := range reqData.RequestList {
			user.RequestList[key] = value
		}
	}

	// process ShadeList if not empty
	if reqData.ShadeList != nil {
		if user.ShadeList == nil {
			user.ShadeList = make(map[string]bool)
		}

		// loop over the ShadeList records and change the user's values accordingly
		// TODO: what if the counterpards shades us already???
		for key, value := range reqData.ShadeList {
			user.ShadeList[key] = value
		}
	}

	// save updated user lists to the database
	if saved := db.SetOne(db.UserCache, userID, user); !saved {
		l.Msg(common.ERR_USER_UPDATE_FAIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("ok, updating user's lists").Status(http.StatusOK).Log().Payload(nil).Write(w)
	return
}

// updateUserOption is the users handler that allows the user to change some attributes of their models.User instance.
//
// @Summary      Update user's option
// @Description  update user's option
// @Tags         users
// @Produce      json
// @Param    	 request body users.updateUserOption.requestData true "new user options' values"
// @Param        userID path string true "ID of the user to update"
// @Success      200  {object}   common.APIResponse
// @Failure      400  {object}   common.APIResponse
// @Failure      403  {object}   common.APIResponse
// @Failure      404  {object}   common.APIResponse
// @Failure      500  {object}   common.APIResponse
// @Router       /users/{userID}/options [patch]
func updateUserOption(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "users")

	type requestData struct {
		UIDarkMode    bool   `json:"dark_mode"`
		LiveMode      bool   `json:"live_mode"`
		LocalTimeMode bool   `json:"local_time_mode"`
		Private       bool   `json:"private"`
		AboutText     string `json:"about_you"`
		WebsiteLink   string `json:"website_link"`
	}

	// skip blank callerID
	if l.CallerID == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// take the param from path
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		l.Msg(common.ERR_USERID_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// check for the possible user's data forgery attempt
	if l.CallerID != userID {
		l.Msg(common.ERR_USER_UPDATE_FOREIGN).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	// fetch the requested user's record form the database
	user, found := db.GetOne(db.UserCache, userID, models.User{})
	if !found {
		l.Msg(common.ERR_USER_NOT_FOUND).Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	//
	// OK, caller seems to be a genuine existing user
	//

	var reqData requestData

	// decode the incoming data
	if err := common.UnmarshalRequestData(r, &reqData); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Error(err).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// patch the nil map
	if user.Options == nil {
		user.Options = make(map[string]bool)
	}

	// toggle dark mode to light mode and vice versa
	if reqData.UIDarkMode != user.UIDarkMode {
		user.UIDarkMode = !user.UIDarkMode
		user.Options["uiDarkMode"] = reqData.UIDarkMode
	}

	// toggle the live mode
	if reqData.LiveMode != user.LiveMode {
		user.LiveMode = !user.LiveMode
		user.Options["liveMode"] = reqData.LiveMode
	}

	// toggle the local time mode
	if reqData.LocalTimeMode != user.LocalTimeMode {
		user.LocalTimeMode = !user.LocalTimeMode
		user.Options["localTimeMode"] = reqData.LocalTimeMode
	}

	// toggle the private mode
	if reqData.Private != user.Private {
		user.Private = !user.Private
		user.Options["private"] = reqData.Private
	}

	// change the about text if present and differs from the current one
	if reqData.AboutText != "" && reqData.AboutText != user.About {
		user.About = reqData.AboutText
	}

	// change the website link if present and differs from the current one
	if reqData.WebsiteLink != "" && reqData.WebsiteLink != user.Web {
		user.Web = reqData.WebsiteLink
	}

	// save the updated user's record back to the database
	if saved := db.SetOne(db.UserCache, userID, user); !saved {
		l.Msg(common.ERR_USER_UPDATE_FAIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("ok, updating user's options").Status(http.StatusOK).Log().Payload(nil).Write(w)
	return
}

// deleteUser is the users handler that processes and deletes given user (oneself) form the database.
//
// @Summary      Delete user
// @Description  delete user
// @Tags         users
// @Produce      json
// @Param        userID path string true "ID of the user to delete"
// @Success      200  {object}   common.APIResponse
// @Failure      400  {object}   common.APIResponse
// @Failure      403  {object}   common.APIResponse
// @Failure      404  {object}   common.APIResponse
// @Failure      500  {object}   common.APIResponse
// @Router       /users/{userID} [delete]
func deleteUser(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "users")

	// skip blank callerID
	if l.CallerID == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// take the param from path
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		l.Msg(common.ERR_USERID_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// check for possible user's data forgery attempt
	if userID != l.CallerID {
		l.Msg(common.ERR_USER_DELETE_FOREIGN).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	// look for such user requested in the database
	if _, found := db.GetOne(db.UserCache, userID, models.User{}); !found {
		l.Msg(common.ERR_USER_NOT_FOUND).Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	// delete requested user record from database
	if deleted := db.DeleteOne(db.UserCache, userID); !deleted {
		l.Msg(common.ERR_USER_DELETE_FAIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	// delete user's subscriptions/registered devices
	if deleted := db.DeleteOne(db.SubscriptionCache, userID); !deleted {
		l.Msg(common.ERR_SUBSCRIPTION_DELETE_FAIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	var void string

	// delete all user's posts and polls, and tokens
	// fetch all posts, polls, and tokens to find mathing ones to delete
	posts, _ := db.GetAll(db.FlowCache, models.Post{})
	polls, _ := db.GetAll(db.PollCache, models.Poll{})
	tokens, _ := db.GetAll(db.TokenCache, void)

	// loop over posts and delete authored ones + linked fungures
	for postID, post := range posts {
		if post.Nickname != userID {
			continue
		}

		// delete such user's post
		if deleted := db.DeleteOne(db.FlowCache, postID); !deleted {
			l.Msg("could not delete deleted user's post (" + postID + ")").Status(http.StatusInternalServerError).Log()
			//continue
		}

		// delete associated image and thumbnail
		if post.Figure != "" {
			// remove the original image's thumbnail
			err := os.Remove("/opt/pix/thumb_" + post.Figure)
			if err != nil {
				l.Msg("error removing a thumbnail: " + err.Error()).Status(http.StatusInternalServerError).Log()
				//continue
			}

			// remove the original image posted
			err = os.Remove("/opt/pix/" + post.Figure)
			if err != nil {
				l.Msg("error removing a original image: " + err.Error()).Status(http.StatusInternalServerError).Log()
				//continue
			}
		}
	}

	// loop over polls and delete authored ones
	for pollID, poll := range polls {
		if poll.Author != userID {
			continue
		}

		// delete such poll
		if deleted := db.DeleteOne(db.PollCache, pollID); !deleted {
			l.Msg("could not delete deleted user's poll (" + pollID + ")").Status(http.StatusInternalServerError).Log()
			continue
		}
	}

	// loop over tokens and delete matching ones
	for tokenHash, token := range tokens {
		if token != userID {
			continue
		}

		// delete such refresh token record
		if deleted := db.DeleteOne(db.TokenCache, tokenHash); !deleted {
			l.Msg("could not delete deleted user's token (" + tokenHash + ")").Status(http.StatusInternalServerError).Log()
			continue
		}
	}

	l.Msg("ok, purging user and their assets").Status(http.StatusOK).Log().Payload(nil).Write(w)
	return
}

// getUserPosts fetches posts only from specified user
//
// @Summary      Get user posts
// @Description  get user posts
// @Tags         users
// @Produce      json
// @Param    	 X-Hide-Replies header string false "hide replies"
// @Param    	 X-Page-No header string true "page number"
// @Param        userID path string true "user's ID for their posts"
// @Success      200  {object}  common.APIResponse{data=users.getUserPosts.responseData}
// @Failure      400  {object}  common.APIResponse
// @Failure      500  {object}  common.APIResponse
// @Router       /users/{userID}/posts [get]
func getUserPosts(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "users")

	type responseData struct {
		Users map[string]models.User `json:"users"`
		Posts map[string]models.Post `json:"posts"`
		Key   string                 `json:"key"`
	}

	// skip blank callerID
	if l.CallerID == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// take the param from path
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		l.Msg(common.ERR_USERID_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// fetch the required X-Page-No header's value
	pageNoString := r.Header.Get(common.HDR_PAGE_NO)
	pageNo, err := strconv.Atoi(pageNoString)
	if err != nil {
		l.Msg(common.ERR_PAGENO_INCORRECT).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// fetch the optional X-Hide-Replies header's value
	hideReplies, err := strconv.ParseBool(r.Header.Get(common.HDR_HIDE_REPLIES))
	if err != nil {
		//l.Msg(common.ERR_HIDE_REPLIES_INVALID).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		//return
		hideReplies = false
	}

	// set the page options
	opts := pages.PageOptions{
		CallerID: l.CallerID,
		PageNo:   pageNo,
		FlowList: nil,

		Flow: pages.FlowOptions{
			HideReplies:  hideReplies,
			Plain:        hideReplies == false,
			UserFlow:     true,
			UserFlowNick: userID,
		},
	}

	// fetch page according to the options and logged user
	pagePtrs := pages.GetOnePage(opts)
	if pagePtrs == (pages.PagePointers{}) || pagePtrs.Posts == nil || pagePtrs.Users == nil || (*pagePtrs.Posts) == nil || (*pagePtrs.Users) == nil {
		l.Msg(common.ERR_PAGE_EXPORT_NIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	// include caller in the user export
	if caller, ok := db.GetOne(db.UserCache, l.CallerID, models.User{}); !ok {
		l.Msg(common.ERR_CALLER_NOT_FOUND).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	} else {
		(*pagePtrs.Users)[l.CallerID] = caller
	}

	// prepare the payload
	pl := &responseData{
		Posts: *pagePtrs.Posts,
		Users: *common.FlushUserData(pagePtrs.Users, l.CallerID),
		Key:   l.CallerID,
	}

	l.Msg("ok, dumping user's flow posts").Status(http.StatusOK).Log().Payload(pl).Write(w)
	return
}

// resetRequestHandler handles the passphrase recovery link generation.
//
// @Summary      Request the passphrase recovery link
// @Description  request the passphrase recovery link
// @Tags         users
// @Accept       json
// @Produce      json
// @Param    	 request body users.resetRequestHandler.requestData true "the e-mail address struct"
// @Success      200  {object}  common.APIResponse
// @Failure      400  {object}  common.APIResponse
// @Failure      404  {object}  common.APIResponse
// @Failure      500  {object}  common.APIResponse
// @Router       /users/passphrase/request [post]
func resetRequestHandler(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "users")

	type requestData struct {
		Email string `json:"email"`
	}

	var reqData requestData

	// decode the incoming data
	if err := common.UnmarshalRequestData(r, &reqData); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	// skip blank e-mail address entered
	if reqData.Email == "" {
		l.Msg(common.ERR_REQUEST_EMAIL_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// preprocess the reqeuest data
	email := strings.ToLower(reqData.Email)
	reqData.Email = email

	// fetch all users to find the matching e-mail address
	users, _ := db.GetAll(db.UserCache, models.User{})

	// loop over users to find matching e-mail address
	var user models.User

	found := false
	for _, usr := range users {
		if strings.ToLower(usr.Email) == reqData.Email {
			found = true
			user = usr
			break
		}
	}

	// e-mail address was not found
	if !found {
		l.Msg(common.ERR_NO_EMAIL_MATCH).Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	// generate new random UUID, aka requestID
	randomID := uuid.New().String()

	// prepare the request data for the database
	dbPayload := models.Request{
		ID:        randomID,
		Nickname:  user.Nickname,
		Email:     email,
		CreatedAt: time.Now(),
	}

	// save new reset request to the database
	if saved := db.SetOne(db.RequestCache, randomID, dbPayload); !saved {
		l.Msg("could not save the UUID to database, try again").Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	// prepare the mail options
	mailPayload := mail.MessagePayload{
		Email: email,
		Type:  "request",
		UUID:  randomID,
	}

	// compose a message to send
	msg, err := mail.ComposeResetMail(mailPayload)
	if err != nil || msg == nil {
		l.Msg(common.ERR_MAIL_COMPOSITION_FAIL).Status(http.StatusInternalServerError).Error(err).Log().Payload(nil).Write(w)
		return
	}

	// send the message
	if err := mail.SendResetMail(msg); err != nil {
		l.Msg(common.ERR_MAIL_NOT_SENT).Status(http.StatusInternalServerError).Error(err).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("the message has been sent, check your e-mail inbox").Status(http.StatusOK).Log().Payload(nil).Write(w)
	return
}

// resetPassphraseHandler handles a new passphrase regeneration.
//
// @Summary      Reset the passphrase
// @Description  reset the passphrase
// @Tags         users
// @Accept       json
// @Produce      json
// @Param    	 request body users.resetPassphraseHandler.requestData true "the e-mail address struct"
// @Success      200  {object}  common.APIResponse
// @Failure      400  {object}  common.APIResponse
// @Failure      404  {object}  common.APIResponse
// @Failure      500  {object}  common.APIResponse
// @Router       /users/passphrase/reset [post]
func resetPassphraseHandler(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "users")

	type requestData struct {
		UUID string `json:"uuid"`
	}

	var reqData requestData

	// decode the incoming data
	if err := common.UnmarshalRequestData(r, &reqData); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Error(err).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// skip blank UUID entered
	if reqData.UUID == "" {
		l.Msg(common.ERR_REQUEST_UUID_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// try the UUID to fetch the request
	request, match := db.GetOne(db.RequestCache, reqData.UUID, models.Request{})
	if !match {
		l.Msg(common.ERR_REQUEST_UUID_INVALID).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	if request.CreatedAt.Before(time.Now().Add(-24 * time.Hour)) {
		// delete the request from database
		if deleted := db.DeleteOne(db.RequestCache, request.Nickname); !deleted {
			l.Msg(common.ERR_REQUEST_DELETE_FAIL).Status(http.StatusInternalServerError).Log()
			//return
		}

		l.Msg(common.ERR_REQUEST_UUID_EXPIRED).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// preprocess the e-mail address
	email := strings.ToLower(request.Email)

	// fetch the user linked to such request
	user, found := db.GetOne(db.UserCache, request.Nickname, models.User{})
	if !found {
		l.Msg(common.ERR_USER_NOT_FOUND).Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	// reset the passphrase - generete a new one (32 chars)
	rand.Seed(time.Now().UnixNano())
	randomPassphrase := helpers.RandSeq(32)
	pepper := os.Getenv("APP_PEPPER")

	// convert new passphrase into the hexadecimal format with pepper added
	passHash := sha512.Sum512([]byte(randomPassphrase + pepper))
	user.PassphraseHex = fmt.Sprintf("%x", passHash)

	// save new passphrase to the database
	if saved := db.SetOne(db.UserCache, user.Nickname, user); !saved {
		l.Msg(common.ERR_PASSPHRASE_UPDATE_FAIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	// set mail options
	mailPayload := mail.MessagePayload{
		Email:      email,
		Type:       "passphrase",
		Passphrase: randomPassphrase,
	}

	// compose a message to send
	msg, err := mail.ComposeResetMail(mailPayload)
	if err != nil || msg == nil {
		l.Msg(common.ERR_MAIL_COMPOSITION_FAIL).Error(err).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	// send the message
	if err := mail.SendResetMail(msg); err != nil {
		l.Msg(common.ERR_MAIL_NOT_SENT).Error(err).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	// delete the request from the request database
	if deleted := db.DeleteOne(db.RequestCache, reqData.UUID); !deleted {
		l.Msg(common.ERR_REQUEST_DELETE_FAIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("the message has been sent, check your e-mail inbox").Status(http.StatusOK).Log().Payload(nil).Write(w)
	return
}

// postUserAvatar is a handler function to update user's avatar directly in the app.
//
// @Summary      Post user's avatar
// @Description  post user's avatar
// @Tags         users
// @Accept       json
// @Produce      json
// @Param    	 request body users.postUserAvatar.requestData true "new avatar data"
// @Param        userID path string true "user's ID for avatar update"
// @Success      200  {object}  common.APIResponse
// @Failure      400  {object}  common.APIResponse
// @Failure      403  {object}  common.APIResponse
// @Failure      404  {object}  common.APIResponse
// @Failure      500  {object}  common.APIResponse
// @Router       /users/{userID}/avatar [post]
func postUserAvatar(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "users")

	type requestData struct {
		Figure string `json."figure"`
		Data   []byte `json:"data"`
	}

	type responseData struct {
		Key string `json:"key"`
	}

	// skip blank callerID
	if l.CallerID == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// take the param from path
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		l.Msg(common.ERR_USERID_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// check for possible user's data forgery attempt
	if l.CallerID != userID {
		l.Msg(common.ERR_USER_AVATAR_FOREIGN).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	// fetch the caller's user record from database
	user, found := db.GetOne(db.UserCache, userID, models.User{})
	if !found {
		l.Msg(common.ERR_USER_NOT_FOUND).Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	var reqData requestData

	// decode the incoming data
	if err := common.UnmarshalRequestData(r, &reqData); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	// prepare stamps and keys for ???
	timestamp := time.Now()
	key := strconv.FormatInt(timestamp.UnixNano(), 10)

	// check if data are there
	if reqData.Data == nil || reqData.Figure == "" {
		l.Msg(common.ERR_MISSING_IMG_CONTENT).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// preprocess the image name and extension
	fileExplode := strings.Split(reqData.Figure, ".")
	extension := fileExplode[len(fileExplode)-1]

	// content for the future user's AvatarURL field update
	content := key + "." + extension

	//
	// use image magic
	//

	// decode the input byte stream into an image.Image object
	img, format, err := image.DecodeImage(&reqData.Data, extension)
	if err != nil {
		l.Msg(common.ERR_IMG_DECODE_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	// fix the image's orientation in decoded image
	img, err = image.FixOrientation(img, &reqData.Data)
	if err != nil {
		l.Msg(common.ERR_IMG_ORIENTATION_FAIL).Status(http.StatusInternalServerError).Error(err).Log().Payload(nil).Write(w)
		return
	}

	// generate thumbanils from the cropped image
	thumbImg := image.ResizeImage(image.CropToSquare(img), 200)

	// encode the thumbnail back to []byte
	thumbImgData, err := image.EncodeImage(&thumbImg, format)
	if err != nil {
		l.Msg(common.ERR_IMG_THUMBNAIL_FAIL).Status(http.StatusInternalServerError).Error(err).Log().Payload(nil).Write(w)
		return
	}

	// write the thumbnail byte stream to a file
	if err := os.WriteFile("/opt/pix/thumb_"+content, *thumbImgData, 0600); err != nil {
		l.Msg(common.ERR_IMG_SAVE_FILE_FAIL).Status(http.StatusInternalServerError).Error(err).Log().Payload(nil).Write(w)
		return
	}

	// null the pointers and contents
	*thumbImgData = []byte{}

	// null the request data
	reqData.Figure = content
	reqData.Data = []byte{}

	// update user's avatar reference
	user.AvatarURL = "/web/pix/thumb_" + content

	// save the updated user record bach to the database
	if saved := db.SetOne(db.UserCache, userID, user); !saved {
		l.Msg(common.ERR_USER_UPDATE_FAIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	// prepare the payload
	pl := &responseData{
		Key: content,
	}

	l.Msg("ok, updating user's avatar").Status(http.StatusOK).Log().Payload(pl).Write(w)
	return
}
