package users

import (
	"crypto/sha512"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"go.savla.dev/littr/pkg/backend/common"
	"go.savla.dev/littr/pkg/backend/db"
	"go.savla.dev/littr/pkg/backend/posts"
	"go.savla.dev/littr/pkg/helpers"
	"go.savla.dev/littr/pkg/models"

	chi "github.com/go-chi/chi/v5"
	mail "github.com/wneessen/go-mail"
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

	// flush email addresses
	for key, user := range users {
		if key == caller {
			continue
		}
		user.Email = ""
		users[key] = user
	}

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

	email := strings.ToLower(user.Email)
	user.Email = email
	user.LastActiveTime = time.Now()

	if saved := db.SetOne(db.UserCache, user.Nickname, user); !saved {
		resp.Message = "cannot save new user"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	//resp.Users[user.Nickname] = user

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
// @Produce      json
// @Success      200  {object}   common.Response
// @Failure      400  {object}   common.Response
// @Failure      404  {object}   common.Response
// @Failure      409  {object}   common.Response
// @Router       /users/{nickname} [put]
func updateUser(w http.ResponseWriter, r *http.Request) {
	resp := common.Response{}
	l := common.NewLogger(r, "users")

	var user models.User

	if err := common.UnmarshalRequestData(r, &user); err != nil {
		resp.Message = "input read error: " + err.Error()
		resp.Code = http.StatusBadRequest

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

	if deleted := db.DeleteOne(db.UserCache, key); !deleted {
		resp.Message = "error deleting:" + key
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	// delete all user's posts and polls
	posts, _ := db.GetAll(db.FlowCache, models.Post{})
	polls, _ := db.GetAll(db.PollCache, models.Poll{})

	for id, post := range posts {
		if post.Nickname == key {
			db.DeleteOne(db.FlowCache, id)
		}
	}

	for id, poll := range polls {
		if poll.Author == key {
			db.DeleteOne(db.PollCache, id)
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

	// flush email addresses
	for key, user := range uExport {
		if key == callerID {
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
	m := mail.NewMsg()
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
	m.SetBodyString(mail.TypeTextPlain, "Someone requested the password reset for the account linked to this e-mail. \n\nNew password:\n\n"+random+"\n\nPlease change your password as soon as possible after a new log-in.")

	port, err := strconv.Atoi(os.Getenv("MAIL_PORT"))
	if err != nil {
		resp.Message = "backend error: cannot convert MAIL_PORT to int: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message+err.Error(), resp.Code)
		resp.Write(w)
		return
	}

	c, err := mail.NewClient(os.Getenv("MAIL_HOST"), mail.WithPort(port), mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(os.Getenv("MAIL_SASL_USR")), mail.WithPassword(os.Getenv("MAIL_SASL_PWD")))
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

// togglePrivateMode is the users' pkg handler to toggle the private status/mode of such user who requested this.
//
// @Summary      Toggle the private mode
// @Description  toggle the private mode
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}  common.Response
// @Failure      403  {object}  common.Response
// @Failure      404  {object}  common.Response
// @Failure      500  {object}  common.Response
// @Router       /users/{nickname}/private [patch]
func togglePrivateMode(w http.ResponseWriter, r *http.Request) {
	resp := common.Response{}
	l := common.NewLogger(r, "users")

	nick := chi.URLParam(r, "nickname")
	caller, _ := r.Context().Value("nickname").(string)

	// disallow unauthorized access
	if nick != caller {
		resp.Message = "access denied"
		resp.Code = http.StatusForbidden

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

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
	user.Private = !user.Private

	if saved := db.SetOne(db.UserCache, nick, user); !saved {
		resp.Message = "backend error: cannot update the user"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	resp.Message = "ok, private mode toggled"
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
