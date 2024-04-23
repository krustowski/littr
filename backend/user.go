package backend

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"go.savla.dev/littr/config"
	"go.savla.dev/littr/models"
)

func getOneUser(w http.ResponseWriter, r *http.Request) {}

func getUsers(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := NewLogger(r, "users")
	stats := make(map[string]userStat)

	caller, _ := r.Context().Value("nickname").(string)
	uuid := r.Header.Get("X-API-Caller-ID")

	// fetch all data for the calculations
	users, _ := getAll(UserCache, models.User{})
	posts, _ := getAll(FlowCache, models.Post{})
	devs, _ := getOne(SubscriptionCache, caller, []models.Device{})

	// check the subscription
	devSubscribed := false
	for _, dev := range devs {
		if dev.UUID == uuid {
			devSubscribed = true
			break
		}
	}

	// assign the result of looping through devices
	resp.Subscribed = devSubscribed

	for _, post := range posts {
		nick := post.Nickname

		var stat userStat
		var found bool
		if stat, found = stats[nick]; !found {
			stat = userStat{}
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

func addNewUser(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := NewLogger(r, "users")

	var user models.User

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		resp.Message = "backend error: cannot read input stream: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	err = json.Unmarshal(reqBody, &user)
	if err != nil {
		resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	if _, found := getOne(UserCache, user.Nickname, models.User{}); found {
		resp.Message = "user already exists"
		resp.Code = http.StatusConflict

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	email := strings.ToLower(user.Email)
	user.Email = email
	user.LastActiveTime = time.Now()

	if saved := setOne(UserCache, user.Nickname, user); !saved {
		resp.Message = "backend error: cannot save new user"
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

func updateUser(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := NewLogger(r, "users")

	var user models.User

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		resp.Message = "backend error: cannot read input stream: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

	err = json.Unmarshal(data, &user)
	if err != nil {
		resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	if _, found := getOne(UserCache, user.Nickname, models.User{}); !found {
		resp.Message = "user not found"
		resp.Code = http.StatusNotFound

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	if saved := setOne(UserCache, user.Nickname, user); !saved {
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

func deleteUser(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := NewLogger(r, "users")

	key, _ := r.Context().Value("nickname").(string)

	if _, found := getOne(UserCache, key, models.User{}); !found {
		resp.Message = "user nout found: " + key
		resp.Code = http.StatusNotFound

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	if deleted := deleteOne(UserCache, key); !deleted {
		resp.Message = "error deleting:" + key
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	// delete all user's posts and polls
	posts, _ := getAll(FlowCache, models.Post{})
	polls, _ := getAll(PollCache, models.Poll{})

	for id, post := range posts {
		if post.Nickname == key {
			deleteOne(FlowCache, id)
		}
	}

	for id, poll := range polls {
		if poll.Author == key {
			deleteOne(PollCache, id)
		}
	}

	resp.Message = "ok, deleting user: " + key
	resp.Code = http.StatusOK

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
}
