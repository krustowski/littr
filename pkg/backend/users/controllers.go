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

	"go.savla.dev/littr/configs"
	"go.savla.dev/littr/pkg/backend/common"
	"go.savla.dev/littr/pkg/backend/db"
	"go.savla.dev/littr/pkg/backend/posts"
	"go.savla.dev/littr/pkg/helpers"
	"go.savla.dev/littr/pkg/models"

	chi "github.com/go-chi/chi/v5"
	gomail "github.com/wneessen/go-mail"
)

// getUsers is the users handler that processes and returns existing users list.
//
// @Summary      Get a list of users
// @Description  get a list of users
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}   common.Response
// @Router       /users/ [get]
func getUsers(w http.ResponseWriter, r *http.Request) {
	resp := common.Response{}
	l := common.NewLogger(r, "users")
	stats := make(map[string]models.UserStat)

	caller, _ := r.Context().Value("nickname").(string)
	uuid := r.Header.Get("X-API-Caller-ID")

	// fetch all data for the calculations
	users, _ := db.GetAll(db.UserCache, models.User{})
	posts, _ := db.GetAll(db.FlowCache, models.Post{})
	devs, _ := db.GetOne(db.SubscriptionCache, caller, []models.Device{})

	// check the subscription
	//devSubscribed := false
	var devTags []string = nil
	for _, dev := range devs {
		if dev.UUID == uuid {
			devTags = dev.Tags
			//devSubscribed = true
			break
		}
	}

	// assign the result of looping through devices
	if helpers.Contains(devTags, "reply") {
		resp.Subscription.Replies = true
	}
	if helpers.Contains(devTags, "mention") {
		resp.Subscription.Mentions = true
	}

	for _, post := range posts {
		nick := post.Nickname

		var stat models.UserStat
		var found bool
		if stat, found = stats[nick]; !found {
			stat = models.UserStat{}
		}

		stat.PostCount++
		stats[nick] = stat
	}

	// calculate the users stats
	for nick, user := range users {
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

	// flush unwanted properties
	for key, user := range users {
		user.Passphrase = ""
		user.PassphraseHex = ""
		user.Email = ""

		if user.Nickname != caller {
			user.FlowList = nil
			user.ShadeList = nil
			user.RequestList = nil
		}

		users[key] = user
	}

	resp.Message = "ok, dumping users"
	resp.Code = http.StatusOK
	resp.Users = users
	resp.UserStats = stats
	resp.Key = caller
	resp.PublicKey = os.Getenv("VAPID_PUB_KEY")
	resp.Devices = devs

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
}

// getOneUser is the users handler that processes and returns existing user's details.
//
// @Summary      Get the user's details
// @Description  get the user's details
// @Tags         users
// @Deprecated
// @Accept       json
// @Produce      json
// @Success      200  {object}   common.Response
// @Failure      400  {object}   common.Response
// @Failure      404  {object}   common.Response
// @Failure      409  {object}   common.Response
// @Router       /users/{nickname} [get]
func getOneUser(w http.ResponseWriter, r *http.Request) {}

// addNewUser is the users handler that processes input and creates a new user.
//
// @Summary      Add new user
// @Description  add new user
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}   common.Response
// @Failure      400  {object}   common.Response
// @Failure      404  {object}   common.Response
// @Failure      409  {object}   common.Response
// @Router       /users/ [post]
func addNewUser(w http.ResponseWriter, r *http.Request) {
	resp := common.Response{}
	l := common.NewLogger(r, "users")

	if !configs.REGISTRATION_ENABLED {
		resp.Message = "registration disallowed at the moment"
		resp.Code = http.StatusForbidden

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	var user models.User

	if err := common.UnmarshalRequestData(r, &user); err != nil {
		resp.Message = "input read error: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	if _, found := db.GetOne(db.UserCache, user.Nickname, models.User{}); found {
		l.Println(
			"user already exists",
			http.StatusConflict,
		)
		resp.Write(w)
		return
	}

	// https://stackoverflow.com/a/38554480
	if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(user.Nickname) {
		resp.Message = "nickname can only have chars a-z, A-Z and numbers"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	email := strings.ToLower(user.Email)
	user.Email = email

	// validate e-mail struct
	// https://stackoverflow.com/a/66624104
	if _, err := mail.ParseAddress(email); err != nil {
		resp.Message = "e-mail address has wrong format"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	user.LastActiveTime = time.Now()

	if saved := db.SetOne(db.UserCache, user.Nickname, user); !saved {
		resp.Message = "cannot save new user"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
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
		resp.Message = "backend error: cannot create a new post about new user adition"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	resp.Message = "ok, adding user"
	resp.Code = http.StatusCreated

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
}

// updateUser is the users handler function that processes and updates given user in the database.
//
// @Summary      Update the user's details
// @Description  update the user's details
// @Tags         users
// @Accept       json
// @Deprecated
// @Produce      json
// @Success      200  {object}   common.Response
// @Failure      400  {object}   common.Response
// @Failure      404  {object}   common.Response
// @Failure      409  {object}   common.Response
// @Router       /users/{nickname} [put]
func updateUser(w http.ResponseWriter, r *http.Request) {
	resp := common.Response{}
	l := common.NewLogger(r, "users")
	callerID, _ := r.Context().Value("nickname").(string)

	var user models.User

	if err := common.UnmarshalRequestData(r, &user); err != nil {
		resp.Message = "input read error: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	if user.Nickname != callerID {
		resp.Message = "one can update theirs account only"
		resp.Code = http.StatusForbidden

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	if _, found := db.GetOne(db.UserCache, user.Nickname, models.User{}); !found {
		resp.Message = "user not found"
		resp.Code = http.StatusNotFound

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	if saved := db.SetOne(db.UserCache, user.Nickname, user); !saved {
		resp.Message = "backend error: cannot update the user"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	resp.Message = "ok, user updated"
	resp.Code = http.StatusCreated

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
}

// updateUserPassphrase is the users handler that allows the user to change their passphrase.
//
// @Summary      Update user's passphrase
// @Description  update user's passphrase
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}   common.Response
// @Failure      403  {object}   common.Response
// @Failure      404  {object}   common.Response
// @Failure      409  {object}   common.Response
// @Failure      500  {object}   common.Response
// @Router       /users/{nickname}/passphrase [patch]
func updateUserPassphrase(w http.ResponseWriter, r *http.Request) {
	resp := common.Response{}
	l := common.NewLogger(r, "users")

	callerID, _ := r.Context().Value("nickname").(string)
	nick := chi.URLParam(r, "nickname")

	if callerID != nick {
		resp.Message = "access denied"
		resp.Code = http.StatusForbidden

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	user, found := db.GetOne(db.UserCache, callerID, models.User{})
	if !found {
		resp.Message = "user nout found: " + callerID
		resp.Code = http.StatusNotFound

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	var options struct {
		PassphraseHex    string `json:"passphrase_hex"`
		OldPassphraseHex string `json:"old_passphrase_hex"`
	}

	if err := common.UnmarshalRequestData(r, &options); err != nil {
		resp.Message = "input read error: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	user.PassphraseHex = options.PassphraseHex

	if saved := db.SetOne(db.UserCache, user.Nickname, user); !saved {
		resp.Message = "backend error: cannot update the user"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}
}

// updateUserOption is the users handler that allows the user to change some attributes of their models.User instance.
//
// @Summary      Update user's option
// @Description  update user's option
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}   common.Response
// @Failure      403  {object}   common.Response
// @Failure      404  {object}   common.Response
// @Failure      409  {object}   common.Response
// @Failure      500  {object}   common.Response
// @Router       /users/{nickname}/options [patch]
func updateUserOption(w http.ResponseWriter, r *http.Request) {
	resp := common.Response{}
	l := common.NewLogger(r, "users")

	callerID, _ := r.Context().Value("nickname").(string)
	nick := chi.URLParam(r, "nickname")

	if callerID != nick {
		resp.Message = "access denied"
		resp.Code = http.StatusForbidden

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	user, found := db.GetOne(db.UserCache, callerID, models.User{})
	if !found {
		resp.Message = "user nout found: " + callerID
		resp.Code = http.StatusNotFound

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
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
		resp.Message = "input read error: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
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
		resp.Message = "backend error: cannot update the user"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}
}

// deleteUser is the users handler that processes and deletes given user (oneself) form the database.
//
// @Summary      Delete user
// @Description  delete user
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}   common.Response
// @Failure      404  {object}   common.Response
// @Failure      409  {object}   common.Response
// @Failure      500  {object}   common.Response
// @Router       /users/{nickname} [delete]
func deleteUser(w http.ResponseWriter, r *http.Request) {
	resp := common.Response{}
	l := common.NewLogger(r, "users")

	key, _ := r.Context().Value("nickname").(string)

	if _, found := db.GetOne(db.UserCache, key, models.User{}); !found {
		resp.Message = "user nout found: " + key
		resp.Code = http.StatusNotFound

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	// delete user
	if deleted := db.DeleteOne(db.UserCache, key); !deleted {
		resp.Message = "error deleting from user cache:" + key
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	// delete user's subscriptions
	if deleted := db.DeleteOne(db.SubscriptionCache, key); !deleted {
		resp.Message = "error deleting from subscription cache:" + key
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	void := ""

	// delete all user's posts and polls, and tokens
	posts, _ := db.GetAll(db.FlowCache, models.Post{})
	polls, _ := db.GetAll(db.PollCache, models.Poll{})
	tokens, _ := db.GetAll(db.TokenCache, void)

	// loop over posts and delete authored ones + linked fungures
	for id, post := range posts {
		if post.Nickname == key {
			db.DeleteOne(db.FlowCache, id)

			// delete associated image and thumbnail
			if post.Figure != "" {
				err := os.Remove("/opt/pix/thumb_" + post.Figure)
				if err != nil {
					// nasty bypass
					continue
				}

				err = os.Remove("/opt/pix/" + post.Figure)
				if err != nil {
					// nasty bypass
					continue
				}
			}
		}
	}

	// loop over polls and delete authored ones
	for id, poll := range polls {
		if poll.Author == key {
			db.DeleteOne(db.PollCache, id)
		}
	}

	// loop over tokens and delete matching ones
	for id, tok := range tokens {
		if tok == key {
			db.DeleteOne(db.TokenCache, id)
		}
	}

	resp.Message = "ok, deleting user: " + key
	resp.Code = http.StatusOK

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
}

// getUserPosts fetches posts only from specified user
//
// @Summary      Get user posts
// @Description  get user posts
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}  common.Response
// @Failure      400  {object}  common.Response
// @Router       /users/{nickname}/posts [get]
func getUserPosts(w http.ResponseWriter, r *http.Request) {
	resp := common.Response{}
	l := common.NewLogger(r, "users")
	callerID, _ := r.Context().Value("nickname").(string)

	// parse the URI's path
	// check if diff page has been requested
	nick := chi.URLParam(r, "nickname")

	pageNo := 0

	pageNoString := r.Header.Get("X-Flow-Page-No")
	page, err := strconv.Atoi(pageNoString)
	if err != nil {
		resp.Message = "page No has to be specified as integer/number"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
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
		resp.Message = "error while requesting more page, one exported map is nil!"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	// hack: include caller's models.User struct
	if caller, ok := db.GetOne(db.UserCache, callerID, models.User{}); !ok {
		resp.Message = "cannot fetch such callerID-named user"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	} else {
		uExport[callerID] = caller
	}

	// flush email addresses
	for key, user := range uExport {
		user.Passphrase = ""
		user.PassphraseHex = ""

		if key == callerID {
			uExport[key] = user
			continue
		}
		user.Email = ""

		uExport[key] = user
	}

	resp.Users = uExport
	resp.Posts = pExport

	resp.Key = callerID

	resp.Message = "ok, dumping user's flow posts"
	resp.Code = http.StatusOK

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
}

// resetHandler poerforms the actual passphrase regeneration and retrieval.
//
// @Summary      Reset the passphrase
// @Description  reset the passphrase
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}  common.Response
// @Failure      400  {object}  common.Response
// @Failure      404  {object}  common.Response
// @Failure      500  {object}  common.Response
// @Router       /users/passphrase [patch]
func resetHandler(w http.ResponseWriter, r *http.Request) {
	resp := common.Response{}
	l := common.NewLogger(r, "users")

	fetch := struct {
		Email      string   `json:"email"`
		Passphrase string   `json:"passphrase"`
		Tags       []string `json:"tags"`
	}{}

	if err := common.UnmarshalRequestData(r, &fetch); err != nil {
		resp.Message = "input read error: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
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
		resp.Message = "backend error: matching user not found"
		resp.Code = http.StatusNotFound

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	rand.Seed(time.Now().UnixNano())
	random := helpers.RandSeq(16)
	pepper := os.Getenv("APP_PEPPER")

	passHash := sha512.Sum512([]byte(random + pepper))
	user.PassphraseHex = fmt.Sprintf("%x", passHash)

	if saved := db.SetOne(db.UserCache, user.Nickname, user); !saved {
		resp.Message = "backend error: cannot update user in database"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	//email := user.Email

	// compose a mail
	m := gomail.NewMsg()
	if err := m.From(os.Getenv("VAPID_SUBSCRIBER")); err != nil {
		resp.Message = "backend error: failed to set From address: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message+err.Error(), resp.Code)
		resp.Write(w)
		return
	}

	if err := m.To(email); err != nil {
		resp.Message = "backend error: failed to set To address: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message+err.Error(), resp.Code)
		resp.Write(w)
		return
	}

	m.Subject("Lost password recovery")
	m.SetBodyString(gomail.TypeTextPlain, "Someone requested the password reset for the account linked to this e-mail. \n\nNew password:\n\n"+random+"\n\nPlease change your password as soon as possible after a new log-in.")

	port, err := strconv.Atoi(os.Getenv("MAIL_PORT"))
	if err != nil {
		resp.Message = "backend error: cannot convert MAIL_PORT to int: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message+err.Error(), resp.Code)
		resp.Write(w)
		return
	}

	c, err := gomail.NewClient(os.Getenv("MAIL_HOST"), gomail.WithPort(port), gomail.WithSMTPAuth(gomail.SMTPAuthPlain),
		gomail.WithUsername(os.Getenv("MAIL_SASL_USR")), gomail.WithPassword(os.Getenv("MAIL_SASL_PWD")))
	if err != nil {
		resp.Message = "backend error: failed to create mail client: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message+err.Error(), resp.Code)
		resp.Write(w)
		return
	}

	//c.SetTLSPolicy(mail.TLSOpportunistic)

	if err := c.DialAndSend(m); err != nil {
		resp.Message = "backend error: failed to sent e-mail: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message+err.Error(), resp.Code)
		resp.Write(w)
		return
	}

	resp.Message = "reset e-mail was rent"
	resp.Code = http.StatusOK

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
	return
}

// addToRequestList is a handler function to add an user to the request list of the private account called as {nickname}.
//
// @Summary      Add to the request list
// @Description  add to the request list
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}  common.Response
// @Failure      400  {object}  common.Response
// @Failure      404  {object}  common.Response
// @Failure      500  {object}  common.Response
// @Router       /users/{nickname}/request [post]
func addToRequestList(w http.ResponseWriter, r *http.Request) {
	resp := common.Response{}
	l := common.NewLogger(r, "users")

	nick := chi.URLParam(r, "nickname")
	caller, _ := r.Context().Value("nickname").(string)

	var user models.User
	var found bool

	if user, found = db.GetOne(db.UserCache, nick, models.User{}); !found {
		resp.Message = "user not found"
		resp.Code = http.StatusNotFound

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	// toggle the status
	if user.RequestList == nil {
		user.RequestList = make(map[string]bool)
	}
	user.RequestList[caller] = true

	if saved := db.SetOne(db.UserCache, nick, user); !saved {
		resp.Message = "backend error: cannot update the user"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	resp.Message = "ok, request addeed to the reqeust list"
	resp.Code = http.StatusOK

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
	return
}

// removeFromRequestList is a handler function to remove an user from the request list of the private account called as {nickname}.
//
// @Summary      Remove from the request list
// @Description  remove from the request list
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}  common.Response
// @Failure      400  {object}  common.Response
// @Failure      404  {object}  common.Response
// @Failure      500  {object}  common.Response
// @Router       /users/{nickname}/request [delete]
func removeFromRequestList(w http.ResponseWriter, r *http.Request) {
	resp := common.Response{}
	l := common.NewLogger(r, "users")

	nick := chi.URLParam(r, "nickname")
	caller, _ := r.Context().Value("nickname").(string)

	var user models.User
	var found bool

	if user, found = db.GetOne(db.UserCache, nick, models.User{}); !found {
		resp.Message = "user not found"
		resp.Code = http.StatusNotFound

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	// toggle the status
	if user.RequestList == nil {
		user.RequestList = make(map[string]bool)
	}
	user.RequestList[caller] = false

	if saved := db.SetOne(db.UserCache, nick, user); !saved {
		resp.Message = "backend error: cannot update the user"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	resp.Message = "ok, request removed from the reqeust list"
	resp.Code = http.StatusOK

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
	return
}

// postUsersAvatar is a handler function to update user's avatar directly in the app.
//
// @Summary      Post user's avatar
// @Description  post user's avatar
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}  common.Response
// @Failure      400  {object}  common.Response
// @Failure      404  {object}  common.Response
// @Failure      500  {object}  common.Response
// @Router       /users/{nickname}/avatar [post]
func postUsersAvatar(w http.ResponseWriter, r *http.Request) {
	resp := common.Response{}
	l := common.NewLogger(r, "users")

	nick := chi.URLParam(r, "nickname")
	caller, _ := r.Context().Value("nickname").(string)

	var user models.User
	var found bool

	if caller != nick {
		resp.Message = "bad request (access denied)"
		resp.Code = http.StatusForbidden

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	if user, found = db.GetOne(db.UserCache, nick, models.User{}); !found {
		resp.Message = "user not found"
		resp.Code = http.StatusNotFound

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	fetch := struct {
		Figure string `json."figure"`
		Data   []byte `json:"data"`
	}{}

	if err := common.UnmarshalRequestData(r, &fetch); err != nil {
		resp.Message = "input read error: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
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

		// upload to local storage
		//if err := os.WriteFile("/opt/pix/"+content, post.Data, 0600); err != nil {
		if err := os.WriteFile("/opt/pix/"+content, fetch.Data, 0600); err != nil {
			resp.Message = "backend error: couldn't save a figure to a file: " + err.Error()
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			resp.Write(w)
			return
		}

		// generate thumbanils
		if err := posts.GenThumbnails("/opt/pix/"+content, "/opt/pix/thumb_"+content); err != nil {
			resp.Message = "backend error: cannot generate the image thumbnail"
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			resp.Write(w)
			return
		}

		fetch.Figure = content
		fetch.Data = []byte{}
	}

	user.AvatarURL = "/web/pix/thumb_" + content

	if saved := db.SetOne(db.UserCache, nick, user); !saved {
		resp.Message = "backend error: cannot save new avatar"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	resp.Message = "ok, updating user's avatar"
	resp.Code = http.StatusOK
	resp.Key = content

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
	return
}
