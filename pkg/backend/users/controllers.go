package users

import (
	"crypto/sha512"
	"fmt"
	"math/rand"
	"net/http"
	"net/mail"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.vxn.dev/littr/configs"
	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/backend/db"
	"go.vxn.dev/littr/pkg/backend/image"
	"go.vxn.dev/littr/pkg/backend/pages"
	"go.vxn.dev/littr/pkg/backend/posts"
	"go.vxn.dev/littr/pkg/helpers"
	"go.vxn.dev/littr/pkg/models"

	chi "github.com/go-chi/chi/v5"
	uuid "github.com/google/uuid"
)

type msgPayload struct {
	Email      string
	Type       string
	UUID       string
	Passphrase string
}

// getUsers is the users handler that processes and returns existing users list.
//
// @Summary      Get a list of users
// @Description  get a list of users
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}   common.APIResponse
// @Router       /users [get]
func getUsers(w http.ResponseWriter, r *http.Request) {
	stats := make(map[string]models.UserStat)

	type payload struct {
		Users     map[string]models.User     `json:"users"`
		User      models.User                `json:"user"`
		UserStats map[string]models.UserStat `json:"user_stats"`
	}

	callerID, _ := r.Context().Value("nickname").(string)

	l := common.NewLogger(r, "users")
	pl := payload{}

	// check the callerID record in the database
	_, ok := db.GetOne(db.UserCache, callerID, models.User{})
	if !ok {
		l.Msg("this user does not exist in the database").Status(http.StatusNotFound).Log().Payload(&pl).Write(w)
		return
	}

	pageNoString := r.Header.Get("X-Page-No")
	pageNo, err := strconv.Atoi(pageNoString)
	if err != nil {
		l.Msg("pageNo has to be specified as integer/number").Status(http.StatusBadRequest).Log().Payload(&pl).Write(w)
		return
	}

	opts := pages.PageOptions{
		CallerID: callerID,
		PageNo:   pageNo,
		FlowList: nil,

		Users: pages.UserOptions{
			Plain: true,
		},
	}

	// fetch one (1) users page
	pagePtrs := pages.GetOnePage(opts)
	if pagePtrs == (pages.PagePointers{}) || pagePtrs.Users == nil || (*pagePtrs.Users) == nil {
		l.Msg("cannot request more pages, one export map is nil!").Status(http.StatusInternalServerError).Log().Payload(&pl).Write(w)
		return
	}

	//(*pagePtrs.Users)[callerID] = caller

	// fetch all data for the calculations
	//users, _ := db.GetAll(db.UserCache, models.User{})
	posts, _ := db.GetAll(db.FlowCache, models.Post{})

	// calculate the post count
	for _, post := range posts {
		nick := post.Nickname

		stat, found := stats[nick]
		if !found {
			stat = models.UserStat{}
		}

		stat.PostCount++
		stats[nick] = stat
	}

	// calculate the flower/follower count
	for nick, user := range *pagePtrs.Users {
		flowList := user.FlowList
		if flowList == nil {
			continue
		}

		for key, state := range flowList {
			if state && key != nick {
				stat := stats[key]
				stat.FlowerCount++
				stats[key] = stat
			}
		}
	}

	// flush unwanted properties in users map
	pl.Users = *common.FlushUserData(pagePtrs.Users, callerID)
	pl.User = pl.Users[callerID]
	pl.UserStats = stats

	l.Msg("ok, dumping users").Status(http.StatusOK).Log().Payload(&pl).Write(w)
	return
}

// getOneUser is the users handler that processes and returns existing user's details.
//
// @Summary      Get the user's details
// @Description  get the user's details
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}   common.APIResponse
// @Failure      404  {object}   common.APIResponse
// @Router       /users/caller [get]
func getOneUser(w http.ResponseWriter, r *http.Request) {
	callerID, _ := r.Context().Value("nickname").(string)

	type payload struct {
		User      models.User     `json:"user"`
		Devices   []models.Device `json:"devices"`
		PublicKey string          `json:"public_key"`
	}

	pl := payload{
		PublicKey: os.Getenv("VAPID_PUB_KEY"),
	}

	l := common.NewLogger(r, "users")

	// check the callerID record in the database
	caller, ok := db.GetOne(db.UserCache, callerID, models.User{})
	if !ok {
		l.Msg("this user does not exist in the database").Status(http.StatusNotFound).Log().Payload(&pl).Write(w)
		return

	}

	// fetch user's devices
	devs, _ := db.GetOne(db.SubscriptionCache, callerID, []models.Device{})

	// helper struct for the data flush
	users := make(map[string]models.User)
	users[callerID] = caller
	users = *common.FlushUserData(&users, callerID)

	// return response
	pl.Devices = devs
	pl.User = users[callerID]

	l.Status(http.StatusOK).Msg("ok, returning callerID's record").Log().Payload(&pl).Write(w)
	return
}

// addNewUser is the users handler that processes input and creates a new user.
//
// @Summary      Add new user
// @Description  add new user
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}   common.APIResponse
// @Failure      400  {object}   common.APIResponse
// @Failure      404  {object}   common.APIResponse
// @Failure      409  {object}   common.APIResponse
// @Router       /users/ [post]
func addNewUser(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "users")

	if !configs.REGISTRATION_ENABLED {
		l.Msg("registration is disabled at the moment").Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	var user models.User

	// decode raw bytes
	if err := common.UnmarshalRequestData(r, &user); err != nil {
		l.Msg("could not process input data, try again").Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// block restricted nicknames, use lowercase for comparison
	if helpers.Contains(configs.UserDeletionList, strings.ToLower(user.Nickname)) {
		l.Msg("this nickname is restricted").Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	if _, found := db.GetOne(db.UserCache, user.Nickname, models.User{}); found {
		l.Msg("this nickname has been already taken").Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// https://stackoverflow.com/a/38554480
	if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(user.Nickname) {
		l.Msg("nickname can consist of chars a-z, A-Z and numbers only").Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// check the nick's length limits
	if len(user.Nickname) > 12 || len(user.Nickname) < 3 {
		l.Msg("nickname is too long (>12 chars), or too short (<3 chars)").Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	email := strings.ToLower(user.Email)
	user.Email = email

	// validate e-mail struct
	// https://stackoverflow.com/a/66624104
	if _, err := mail.ParseAddress(email); err != nil {
		l.Msg("e-mail address is of a wrong format").Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// check for the already registred e-mail
	allUsers, _ := db.GetAll(db.UserCache, models.User{})

	for _, u := range allUsers {
		if strings.ToLower(u.Email) == user.Email {
			l.Msg("this e-mail address has been already used").Status(http.StatusConflict).Log().Payload(nil).Write(w)
			return
		}
	}

	// validation end = new user can be added

	user.LastActiveTime = time.Now()
	user.About = "newbie"

	if saved := db.SetOne(db.UserCache, user.Nickname, user); !saved {
		l.Msg("could not save new user, try again").Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	//resp.Users[user.Nickname] = user

	postStamp := time.Now()
	postKey := strconv.FormatInt(postStamp.UnixNano(), 10)

	post := models.Post{
		ID:        postKey,
		Type:      "user",
		Figure:    user.Nickname,
		Nickname:  "system",
		Content:   "new user has been added (" + user.Nickname + ")",
		Timestamp: postStamp,
	}

	if saved := db.SetOne(db.FlowCache, postKey, post); !saved {
		l.Msg("could not create a new post about new registration").Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("ok, adding new user").Status(http.StatusCreated).Log().Payload(nil).Write(w)
	return
}

// updateUser is the users handler function that processes and updates given user in the database.
//
// @Summary      Update the user's details
// @Description  update the user's details
// @Tags         users
// @Accept       json
// @Deprecated
// @Produce      json
// @Success      200  {object}   common.APIResponse
// @Failure      400  {object}   common.APIResponse
// @Failure      404  {object}   common.APIResponse
// @Failure      409  {object}   common.APIResponse
// @Router       /users/{nickname} [put]
func updateUser(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "users")
	callerID, _ := r.Context().Value("nickname").(string)

	var user models.User

	if err := common.UnmarshalRequestData(r, &user); err != nil {
		l.Msg("could not create process input data, try again").Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	if user.Nickname != callerID {
		l.Msg("you can update yours account only").Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	if _, found := db.GetOne(db.UserCache, user.Nickname, models.User{}); !found {
		l.Msg("user not found").Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	if saved := db.SetOne(db.UserCache, user.Nickname, user); !saved {
		l.Msg("could not update the user in database, try again").Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
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
// @Accept       json
// @Produce      json
// @Success      200  {object}   common.APIResponse
// @Failure      403  {object}   common.APIResponse
// @Failure      404  {object}   common.APIResponse
// @Failure      409  {object}   common.APIResponse
// @Failure      500  {object}   common.APIResponse
// @Router       /users/{nickname}/passphrase [patch]
func updateUserPassphrase(w http.ResponseWriter, r *http.Request) {
	callerID, _ := r.Context().Value("nickname").(string)
	nick := chi.URLParam(r, "nickname")

	l := common.NewLogger(r, "users")

	if callerID != nick {
		l.Msg("you can update yours passphrase only").Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	user, found := db.GetOne(db.UserCache, callerID, models.User{})
	if !found {
		l.Msg("user not found in the database").Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	var options struct {
		NewPassphraseHex     string `json:"new_passphrase_hex"`
		CurrentPassphraseHex string `json:"current_passphrase_hex"`
	}

	if err := common.UnmarshalRequestData(r, &options); err != nil {
		l.Msg("could not process input data, try again").Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	if options.NewPassphraseHex == "" || options.CurrentPassphraseHex == "" {
		l.Msg("empty data received, try again").Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	if options.CurrentPassphraseHex != user.PassphraseHex {
		l.Msg("current passhrase is wrong").Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	user.PassphraseHex = options.NewPassphraseHex

	if saved := db.SetOne(db.UserCache, user.Nickname, user); !saved {
		l.Msg("could not update the user in database, try again").Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
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
// @Accept       json
// @Produce      json
// @Success      200  {object}   common.APIResponse
// @Failure      403  {object}   common.APIResponse
// @Failure      404  {object}   common.APIResponse
// @Failure      409  {object}   common.APIResponse
// @Failure      500  {object}   common.APIResponse
// @Router       /users/{nickname}/lists [patch]
func updateUserList(w http.ResponseWriter, r *http.Request) {
	callerID, _ := r.Context().Value("nickname").(string)
	nick := chi.URLParam(r, "nickname")

	l := common.NewLogger(r, "users")

	user, found := db.GetOne(db.UserCache, nick, models.User{})
	if !found {
		l.Msg("user not found in the database").Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	lists := struct {
		FlowList    map[string]bool `json:"flow_list"`
		RequestList map[string]bool `json:"request_list"`
		ShadeList   map[string]bool `json:"shade_list"`
	}{}

	if err := common.UnmarshalRequestData(r, &lists); err != nil {
		l.Msg("could not process input data, try again").Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	// process FlowList if not empty
	if lists.FlowList != nil {
		if user.FlowList == nil {
			user.FlowList = make(map[string]bool)
		}

		for key, value := range lists.FlowList {
			// do not allow to unfollow oneself
			if key == user.Nickname {
				user.FlowList[key] = true
				continue
			}

			if _, found := user.FlowList[key]; found {
				user.FlowList[key] = value
			}

			// check if the user is shaded by the counterpart
			if counterpart, exists := db.GetOne(db.UserCache, key, models.User{}); exists {
				if counterpart.Private && key != callerID {
					// cannot add this user to one's flow, as the following
					// has to be requested and allowed by the counterpart
					continue
				}

				if shaded, found := counterpart.ShadeList[user.Nickname]; !found || (found && !shaded) {
					user.FlowList[key] = value
				}
			}
		}
	}
	user.FlowList["system"] = true

	// process RequestList if not empty
	if lists.RequestList != nil {
		if user.RequestList == nil {
			user.RequestList = make(map[string]bool)
		}

		for key, value := range lists.RequestList {
			user.RequestList[key] = value
		}
	}

	// process ShadeList if not empty
	if lists.ShadeList != nil {
		if user.ShadeList == nil {
			user.ShadeList = make(map[string]bool)
		}

		for key, value := range lists.ShadeList {
			user.ShadeList[key] = value
		}
	}

	if saved := db.SetOne(db.UserCache, nick, user); !saved {
		l.Msg("could not update the user in database, try again").Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
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
// @Accept       json
// @Produce      json
// @Success      200  {object}   common.APIResponse
// @Failure      403  {object}   common.APIResponse
// @Failure      404  {object}   common.APIResponse
// @Failure      409  {object}   common.APIResponse
// @Failure      500  {object}   common.APIResponse
// @Router       /users/{nickname}/options [patch]
func updateUserOption(w http.ResponseWriter, r *http.Request) {
	callerID, _ := r.Context().Value("nickname").(string)
	nick := chi.URLParam(r, "nickname")

	l := common.NewLogger(r, "users")

	if callerID != nick {
		l.Msg("you can update yours data only").Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	user, found := db.GetOne(db.UserCache, callerID, models.User{})
	if !found {
		l.Msg("user not found in the database").Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	var options struct {
		UIDarkMode    bool   `json:"dark_mode"`
		LiveMode      bool   `json:"live_mode"`
		LocalTimeMode bool   `json:"local_time_mode"`
		Private       bool   `json:"private"`
		AboutText     string `json:"about_you"`
		WebsiteLink   string `json:"website_link"`
	}

	if err := common.UnmarshalRequestData(r, &options); err != nil {
		l.Msg("could not process input data, try again").Error(err).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// toggle dark mode to light mode and vice versa
	if options.UIDarkMode != user.UIDarkMode {
		user.UIDarkMode = !user.UIDarkMode
	}

	// toggle the live mode
	if options.LiveMode != user.LiveMode {
		user.LiveMode = !user.LiveMode
	}

	// toggle the local time mode
	if options.LocalTimeMode != user.LocalTimeMode {
		user.LocalTimeMode = !user.LocalTimeMode
	}

	// toggle the private mode
	if options.Private != user.Private {
		user.Private = !user.Private
	}

	// change the about text if present and differs from the current one
	if options.AboutText != "" && options.AboutText != user.About {
		user.About = options.AboutText
	}

	// change the website link if present and differs from the current one
	if options.WebsiteLink != "" && options.WebsiteLink != user.Web {
		user.Web = options.WebsiteLink
	}

	if saved := db.SetOne(db.UserCache, user.Nickname, user); !saved {
		l.Msg("could not update the user in database, try again").Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
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
// @Accept       json
// @Produce      json
// @Success      200  {object}   common.APIResponse
// @Failure      404  {object}   common.APIResponse
// @Failure      409  {object}   common.APIResponse
// @Failure      500  {object}   common.APIResponse
// @Router       /users/{nickname} [delete]
func deleteUser(w http.ResponseWriter, r *http.Request) {
	nick, _ := r.Context().Value("nickname").(string)

	l := common.NewLogger(r, "users")

	if _, found := db.GetOne(db.UserCache, nick, models.User{}); !found {
		l.Msg("user not found in the database").Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	// delete user
	if deleted := db.DeleteOne(db.UserCache, nick); !deleted {
		l.Msg("could not delete the user from user database").Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	// delete user's subscriptions
	if deleted := db.DeleteOne(db.SubscriptionCache, nick); !deleted {
		l.Msg("could not delete the user from subscription database").Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	void := ""

	// delete all user's posts and polls, and tokens
	posts, _ := db.GetAll(db.FlowCache, models.Post{})
	polls, _ := db.GetAll(db.PollCache, models.Poll{})
	tokens, _ := db.GetAll(db.TokenCache, void)

	// loop over posts and delete authored ones + linked fungures
	for id, post := range posts {
		if post.Nickname == nick {
			db.DeleteOne(db.FlowCache, id)

			// delete associated image and thumbnail
			if post.Figure != "" {
				err := os.Remove("/opt/pix/thumb_" + post.Figure)
				if err != nil {
					// TODO catch error
					continue
				}

				err = os.Remove("/opt/pix/" + post.Figure)
				if err != nil {
					// TODO catch error
					continue
				}
			}
		}
	}

	// loop over polls and delete authored ones
	for id, poll := range polls {
		if poll.Author == nick {
			// TODO catch error
			db.DeleteOne(db.PollCache, id)
		}
	}

	// loop over tokens and delete matching ones
	for id, tok := range tokens {
		if tok == nick {
			// TODO catch error
			db.DeleteOne(db.TokenCache, id)
		}
	}

	l.Msg("ok, deleting user " + nick).Status(http.StatusOK).Log().Payload(nil).Write(w)
	return
}

// getUserPosts fetches posts only from specified user
//
// @Summary      Get user posts
// @Description  get user posts
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}  common.APIResponse
// @Failure      400  {object}  common.APIResponse
// @Router       /users/{nickname}/posts [get]
func getUserPosts(w http.ResponseWriter, r *http.Request) {
	callerID, _ := r.Context().Value("nickname").(string)

	l := common.NewLogger(r, "users")

	// parse the URI's path
	// check if diff page has been requested
	nick := chi.URLParam(r, "nickname")

	pageNo := 0

	pageNoString := r.Header.Get("X-Page-No")
	page, err := strconv.Atoi(pageNoString)
	if err != nil {
		l.Msg("pageNo has to be specified as integer/number").Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	pageNo = page

	// mock the flowlist (nasty hack)
	flowList := make(map[string]bool)
	flowList[nick] = true

	opts := posts.PageOptions{
		UserFlow:     true,
		UserFlowNick: nick,
		CallerID:     callerID,
		PageNo:       pageNo,
		FlowList:     flowList,
	}

	// fetch page according to the logged user
	pExport, uExport := posts.GetOnePage(opts)
	if pExport == nil || uExport == nil {
		l.Msg("could not request more pages, one exported map is nil!").Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	// hack: include caller's models.User struct
	if caller, ok := db.GetOne(db.UserCache, callerID, models.User{}); !ok {
		l.Msg("could not get the user from database, try again").Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	} else {
		uExport[callerID] = caller
	}

	pl := struct {
		Users map[string]models.User
		Posts map[string]models.Post
		Key   string
	}{
		Users: *common.FlushUserData(&uExport, callerID),
		Posts: pExport,
		Key:   callerID,
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
// @Success      200  {object}  common.APIResponse
// @Failure      400  {object}  common.APIResponse
// @Failure      404  {object}  common.APIResponse
// @Failure      500  {object}  common.APIResponse
// @Router       /users/passphrase/request [post]
func resetRequestHandler(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "users")

	fetch := struct {
		Email string `json:"email"`
	}{}

	if err := common.UnmarshalRequestData(r, &fetch); err != nil {
		l.Msg("could not process the input data, try again").Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	email := strings.ToLower(fetch.Email)
	fetch.Email = email

	users, _ := db.GetAll(db.UserCache, models.User{})

	// loop over users to find matching e-mail address
	var user models.User

	found := false
	for _, u := range users {
		if strings.ToLower(u.Email) == fetch.Email {
			found = true
			user = u
			break
		}
	}

	if !found {
		l.Msg("no match found in the database").Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	randomID := uuid.New().String()

	dbPayload := models.Request{
		ID:        randomID,
		Nickname:  user.Nickname,
		Email:     email,
		CreatedAt: time.Now(),
	}

	if saved := db.SetOne(db.RequestCache, randomID, dbPayload); !saved {
		l.Msg("could not save the UUID to database, try again").Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	mailPayload := msgPayload{
		Email: email,
		Type:  "request",
		UUID:  randomID,
	}

	// compose a message to send
	msg, err := composeResetMail(mailPayload)
	if err != nil || msg == nil {
		l.Msg("could not compose the mail body, try again").Status(http.StatusInternalServerError).Error(err).Log().Payload(nil).Write(w)
		return
	}

	if err := sendResetMail(msg); err != nil {
		l.Msg("could not send the mail, try again").Status(http.StatusInternalServerError).Error(err).Log().Payload(nil).Write(w)
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
// @Success      200  {object}  common.APIResponse
// @Failure      400  {object}  common.APIResponse
// @Failure      404  {object}  common.APIResponse
// @Failure      500  {object}  common.APIResponse
// @Router       /users/passphrase/reset [post]
func resetPassphraseHandler(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "users")

	fetch := struct {
		UUID string `json:"uuid"`
	}{}

	if err := common.UnmarshalRequestData(r, &fetch); err != nil {
		l.Msg("could not process tbe input data, try again").Error(err).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	if fetch.UUID == "" {
		l.Msg("no UUID has been entered").Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	requestData, match := db.GetOne(db.RequestCache, fetch.UUID, models.Request{})
	if !match {
		l.Msg("invalid UUID inserted").Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	if requestData.CreatedAt.Before(time.Now().Add(-24 * time.Hour)) {
		// delete the request from database
		db.DeleteOne(db.RequestCache, requestData.Nickname)

		l.Msg("entered UUID expired").Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	email := strings.ToLower(requestData.Email)

	user, found := db.GetOne(db.UserCache, requestData.Nickname, models.User{})
	if !found {
		l.Msg("no match has been found in the database").Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	// reset the passphrase - generete a new one (32 chars)
	rand.Seed(time.Now().UnixNano())
	randomPassphrase := helpers.RandSeq(32)
	pepper := os.Getenv("APP_PEPPER")

	// convert new passphrase into the hexadecimal format
	passHash := sha512.Sum512([]byte(randomPassphrase + pepper))
	user.PassphraseHex = fmt.Sprintf("%x", passHash)

	if saved := db.SetOne(db.UserCache, user.Nickname, user); !saved {
		l.Msg("could not update user's passphrase, try again").Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	mailPayload := msgPayload{
		Email:      email,
		Type:       "passphrase",
		Passphrase: randomPassphrase,
	}

	// compose a message to send
	msg, err := composeResetMail(mailPayload)
	if err != nil || msg == nil {
		l.Msg("could not compose the mail body, try again").Error(err).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	if err := sendResetMail(msg); err != nil {
		l.Msg("could not send the mail, try again").Error(err).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	if deleted := db.DeleteOne(db.RequestCache, fetch.UUID); !deleted {
		l.Msg("could not delete the request from database, try again").Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("the message has been sent, check your e-mail inbox").Status(http.StatusOK).Log().Payload(nil).Write(w)
	return
}

// addToRequestList is a handler function to add an user to the request list of the private account called as {nickname}.
//
// @Summary      Add to the request list
// @Description  add to the request list
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}  common.APIResponse
// @Failure      400  {object}  common.APIResponse
// @Failure      404  {object}  common.APIResponse
// @Failure      500  {object}  common.APIResponse
// @Router       /users/{nickname}/request [post]
func addToRequestList(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "users")

	callerID, _ := r.Context().Value("nickname").(string)
	nick := chi.URLParam(r, "nickname")

	var caller models.User
	var found bool

	if caller, found = db.GetOne(db.UserCache, callerID, models.User{}); !found {
		l.Msg("caller not found in the database").Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	var user models.User

	if user, found = db.GetOne(db.UserCache, nick, models.User{}); !found {
		l.Msg("user not found in the database").Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	// patch the nil in RequestList
	if user.RequestList == nil {
		user.RequestList = make(map[string]bool)
	}

	// toggle the status for the user
	user.RequestList[caller.Nickname] = true

	if saved := db.SetOne(db.UserCache, user.Nickname, user); !saved {
		l.Msg("could not update the user in database, try again").Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("ok, request added to the database").Status(http.StatusOK).Log().Payload(nil).Write(w)
	return
}

// removeFromRequestList is a handler function to remove an user from the request list of the private account called as {nickname}.
//
// @Summary      Remove from the request list
// @Description  remove from the request list
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}  common.APIResponse
// @Failure      400  {object}  common.APIResponse
// @Failure      404  {object}  common.APIResponse
// @Failure      500  {object}  common.APIResponse
// @Router       /users/{nickname}/request [delete]
func removeFromRequestList(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "users")

	nick := chi.URLParam(r, "nickname")

	callerID, ok := r.Context().Value("nickname").(string)
	if !ok {
		l.Msg("could not get the caller's name, try again").Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
	}

	var caller models.User
	var found bool

	if caller, found = db.GetOne(db.UserCache, callerID, models.User{}); !found {
		l.Msg("caller not found in the database").Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	var user models.User

	if user, found = db.GetOne(db.UserCache, nick, models.User{}); !found {
		l.Msg("user not found in the database").Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	// patch the nil in RequestList
	if user.RequestList == nil {
		user.RequestList = make(map[string]bool)
	}

	// toggle the status
	user.RequestList[caller.Nickname] = false

	if saved := db.SetOne(db.UserCache, nick, user); !saved {
		l.Msg("could not update the user in database, try again").Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("ok, request has been removed from database").Status(http.StatusOK).Log().Payload(nil).Write(w)
	return
}

// postUsersAvatar is a handler function to update user's avatar directly in the app.
//
// @Summary      Post user's avatar
// @Description  post user's avatar
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}  common.APIResponse
// @Failure      400  {object}  common.APIResponse
// @Failure      404  {object}  common.APIResponse
// @Failure      500  {object}  common.APIResponse
// @Router       /users/{nickname}/avatar [post]
func postUsersAvatar(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "users")

	nick := chi.URLParam(r, "nickname")

	callerID, ok := r.Context().Value("nickname").(string)
	if !ok {
		l.Msg("could not get the caller's name, try again").Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
	}

	var user models.User
	var found bool

	if callerID != nick {
		l.Msg("you can update yours avatar only").Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	if user, found = db.GetOne(db.UserCache, nick, models.User{}); !found {
		l.Msg("user not found in the database, try again or relog").Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	fetch := struct {
		Figure string `json."figure"`
		Data   []byte `json:"data"`
	}{}

	if err := common.UnmarshalRequestData(r, &fetch); err != nil {
		l.Msg("could not process the input data, try again").Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	timestamp := time.Now()
	key := strconv.FormatInt(timestamp.UnixNano(), 10)

	var content string

	// uploadedFigure handling
	if fetch.Data != nil && fetch.Figure != "" {
		fileExplode := strings.Split(fetch.Figure, ".")
		extension := fileExplode[len(fileExplode)-1]

		content = key + "." + extension

		//
		// use image magic
		//
		img, format, err := image.DecodeImage(&fetch.Data, extension)
		if err != nil {
			l.Msg("image: could not decode given byte stream").Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
			return
		}

		// fix the image orientation for decoded image
		img, err = image.FixOrientation(img, &fetch.Data)
		if err != nil {
			l.Msg("image: could not fix the orientation").Status(http.StatusInternalServerError).Error(err).Log().Payload(nil).Write(w)
			return
		}

		// encode cropped image back to []byte
		// re-encode the image to flush EXIF metadata header
		/*croppedImgData, err := image.EncodeImage(squareImg, format)
		if err != nil {
			resp.Message = "backend error: cannot encode image back to byte stream"
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			resp.Write(w)
			return
		}

		// upload to local storage
		//if err := os.WriteFile("/opt/pix/"+content, post.Data, 0600); err != nil {
		if err := os.WriteFile("/opt/pix/"+content, croppedImgData, 0600); err != nil {
			resp.Message = "backend error: couldn't save a figure to a file: " + err.Error()
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			resp.Write(w)
			return
		}*/

		// crop the image
		squareImg := image.CropToSquare(img)

		// generate thumbanils
		thumbImg := image.ResizeImage(squareImg, 200)
		*squareImg = nil

		// encode the thumbnail back to []byte
		thumbImgData, err := image.EncodeImage(&thumbImg, format)
		if err != nil {
			l.Msg("image: could not encode the thumbnail to byte stream").Status(http.StatusInternalServerError).Error(err).Log().Payload(nil).Write(w)
			return
		}

		// write the thumbnail byte stream to a file
		if err := os.WriteFile("/opt/pix/thumb_"+content, *thumbImgData, 0600); err != nil {
			l.Msg("image: could not save the file, try again").Status(http.StatusInternalServerError).Error(err).Log().Payload(nil).Write(w)
			return
		}

		*thumbImgData = []byte{}

		fetch.Figure = content
		fetch.Data = []byte{}
	}

	user.AvatarURL = "/web/pix/thumb_" + content

	if saved := db.SetOne(db.UserCache, nick, user); !saved {
		l.Msg("could not save new avatar to database").Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	pl := struct {
		Key string
	}{
		Key: content,
	}

	l.Msg("ok, updating user's avatar").Status(http.StatusOK).Log().Payload(&pl).Write(w)
	return
}
