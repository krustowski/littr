package backend

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.savla.dev/littr/config"
	"go.savla.dev/littr/models"
)

func AuthHandler(w http.ResponseWriter, r *http.Request) {
	resp := response{}

	// prepare the Logger instance
	l := Logger{
		IPAddress:  r.Header.Get("X-Real-IP"),
		Method:     r.Method,
		Route:      r.URL.String(),
		WorkerName: "auth",
	}

	resp.AuthGranted = false

	switch r.Method {
	case "POST":
		var user models.User

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			resp.Message = "backend error: cannot read input stream"
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message+err.Error(), resp.Code)
			break
		}

		data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

		if err = json.Unmarshal(data, &user); err != nil {
			resp.Message = "backend error: cannot unmarshall request data"
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message+err.Error(), resp.Code)
			break
		}

		l.CallerID = user.Nickname

		// try to authenticate given user
		u, ok := authUser(user)
		if !ok {
			resp.Message = "user not found, or wrong passphrase entered"
			resp.Code = http.StatusBadRequest

			l.Println(resp.Message, resp.Code)
			break
		}

		resp.Users = make(map[string]models.User)
		resp.Users[u.Nickname] = *u
		resp.AuthGranted = ok

		resp.Message = "auth granted"
		resp.Code = http.StatusOK
		//resp.FlowList = u.FlowListdd

		l.Println(resp.Message, resp.Code)
		break

	default:
		resp.Message = "disallowed method"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		break
	}

	resp.Write(w)
}

func DumpHandler(w http.ResponseWriter, r *http.Request) {
	resp := response{}

	// prepare the Logger instance
	l := Logger{
		CallerID: "system",
		//IPAddress:  r.RemoteAddr,
		IPAddress:  r.Header.Get("X-Real-IP"),
		Method:     r.Method,
		Route:      r.URL.String(),
		WorkerName: "dump",
	}

	// check the incoming API token
	token := r.Header.Get("X-App-Token")

	if token == "" {
		resp.Message = "empty token"
		resp.Code = http.StatusUnauthorized

		l.Println(resp.Message, resp.Code)
		return
	}

	if token != os.Getenv("API_TOKEN") {
		resp.Message = "invalid token"
		resp.Code = http.StatusForbidden

		l.Println(resp.Message, resp.Code)
		return
	}

	switch r.Method {
	case "GET":
		DumpData()

		resp.Code = http.StatusOK
		resp.Message = "data dumped successfully"

		l.Println(resp.Message, resp.Code)
		break

	default:
		resp.Message = "disallowed method"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		break
	}

	// dynamic encryption bypass hack --- we need unecrypted JSON for curl (healthcheck)
	if config.EncryptionEnabled {
		//log.Printf("[dump] disabling encryption (was %t)", config.EncryptionEnabled)
		config.EncryptionEnabled = !config.EncryptionEnabled

		resp.Write(w)

		//log.Printf("[dump] enabling encryption (was %t)", config.EncryptionEnabled)
		config.EncryptionEnabled = !config.EncryptionEnabled
	} else {
		resp.Write(w)
	}
}

func FlowHandler(w http.ResponseWriter, r *http.Request) {
	resp := response{}

	// prepare the Logger instance
	l := Logger{
		CallerID:  r.Header.Get("X-API-Caller-ID"),
		IPAddress: r.Header.Get("X-Real-IP"),
		//IPAddress:  r.RemoteAddr,
		Method:     r.Method,
		Route:      r.URL.String(),
		WorkerName: "flow",
	}

	noteUsersActivity(r.Header.Get("X-API-Caller-ID"))

	switch r.Method {
	case "DELETE":
		// remove a post
		var post models.Post

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			resp.Message = "backend error: cannot read input stream: " + err.Error()
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			break
		}

		data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

		if err := json.Unmarshal(data, &post); err != nil {
			resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			break
		}

		//key := strconv.FormatInt(post.Timestamp.UnixNano(), 10)
		key := post.ID

		if deleted := deleteOne(FlowCache, key); !deleted {
			resp.Message = "cannot delete the post"
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			break
		}

		resp.Message = "ok, post removed"
		resp.Code = http.StatusOK

		l.Println(resp.Message, resp.Code)
		break

	case "GET":
		// fetch the flow, ergo post list
		posts, count := getAll(FlowCache, models.Post{})
		//posts, count := getMany(FlowCache, models.Post{}, "", 5, true)

		resp.Message = "ok, dumping posts"
		resp.Code = http.StatusOK
		resp.Posts = posts
		resp.Count = count

		l.Println(resp.Message, resp.Code)
		break

	case "POST":
		// post a new post
		var post models.Post

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			resp.Message = "backend error: cannot read input stream: " + err.Error()
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			break
		}

		data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

		if err := json.Unmarshal(data, &post); err != nil {
			resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			break
		}

		timestamp := time.Now()
		key := strconv.FormatInt(timestamp.UnixNano(), 10)

		post.ID = key
		post.Timestamp = timestamp

		if saved := setOne(FlowCache, key, post); !saved {
			resp.Message = "backend error: cannot save new post (cache error)"
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			break
		}

		posts, _ := getAll(FlowCache, models.Post{})

		resp.Message = "ok, adding new post"
		resp.Code = http.StatusCreated
		resp.Posts = posts

		l.Println(resp.Message, resp.Code)
		break

	case "PUT":
		// edit/update a post
		var post models.Post

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			resp.Message = "backend error: cannot read input stream: " + err.Error()
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			break
		}

		data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

		if err := json.Unmarshal(data, &post); err != nil {
			resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			break
		}

		//key := strconv.FormatInt(post.Timestamp.UnixNano(), 10)
		key := post.ID

		if _, found := getOne(FlowCache, key, models.Post{}); !found {
			resp.Message = "unknown post update requested"
			resp.Code = http.StatusBadRequest

			l.Println(resp.Message, resp.Code)
			break
		}

		if saved := setOne(FlowCache, key, post); !saved {
			resp.Message = "cannot update post"
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			break
		}

		resp.Message = "ok, post updated"
		resp.Code = http.StatusOK

		l.Println(resp.Message, resp.Code)
		break

	default:
		resp.Message = "disallowed method"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		break
	}

	resp.Write(w)
}

func PollsHandler(w http.ResponseWriter, r *http.Request) {
	resp := response{}

	// prepare the Logger instance
	l := Logger{
		CallerID:  r.Header.Get("X-API-Caller-ID"),
		IPAddress: r.Header.Get("X-Real-IP"),
		//IPAddress:  r.RemoteAddr,
		Method:     r.Method,
		Route:      r.URL.String(),
		WorkerName: "polls",
	}

	noteUsersActivity(r.Header.Get("X-API-Caller-ID"))

	switch r.Method {
	case "DELETE":
		var poll models.Poll

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			resp.Message = "backend error: cannot read input stream: " + err.Error()
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			break
		}

		data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

		if err := json.Unmarshal(data, &poll); err != nil {
			resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			break
		}

		key := poll.ID

		if deleted := deleteOne(PollCache, key); !deleted {
			resp.Message = "cannot delete the poll"
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			break
		}

		resp.Message = "ok, poll removed"
		resp.Code = http.StatusOK

		l.Println(resp.Message, resp.Code)
		break

	case "GET":
		polls, _ := getAll(PollCache, models.Poll{})

		resp.Message = "ok, dumping polls"
		resp.Code = http.StatusOK
		resp.Polls = polls

		l.Println(resp.Message, resp.Code)
		break

	case "POST":
		var poll models.Poll

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			resp.Message = "backend error: cannot read input stream: " + err.Error()
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			break
		}

		data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

		if err := json.Unmarshal(data, &poll); err != nil {
			resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			break
		}

		key := poll.ID

		if saved := setOne(PollCache, key, poll); !saved {
			resp.Message = "backend error: cannot save new poll (cache error)"
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			break
		}

		resp.Message = "ok, adding new poll"
		resp.Code = http.StatusCreated

		l.Println(resp.Message, resp.Code)
		break

	case "PUT":
		var poll models.Poll

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			resp.Message = "backend error: cannot read input stream: " + err.Error()
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			break
		}

		data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

		if err := json.Unmarshal(data, &poll); err != nil {
			resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			break
		}

		key := poll.ID

		if _, found := getOne(PollCache, key, models.Poll{}); !found {
			resp.Message = "unknown poll update requested"
			resp.Code = http.StatusBadRequest

			l.Println(resp.Message, resp.Code)
			break
		}

		if saved := setOne(PollCache, key, poll); !saved {
			resp.Message = "cannot update post"
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			break
		}

		resp.Message = "ok, poll updated"
		resp.Code = http.StatusOK

		l.Println(resp.Message, resp.Code)
		break

	default:
		resp.Message = "disallowed method"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		break
	}

	resp.Write(w)
}

func UsersHandler(w http.ResponseWriter, r *http.Request) {
	resp := response{}

	// prepare the Logger instance
	l := Logger{
		CallerID:  r.Header.Get("X-API-Caller-ID"),
		IPAddress: r.Header.Get("X-Real-IP"),
		//IPAddress:  r.RemoteAddr,
		Method:     r.Method,
		Route:      r.URL.String(),
		WorkerName: "users",
	}

	noteUsersActivity(r.Header.Get("X-API-Caller-ID"))

	switch r.Method {
	case "DELETE":
		// remove an user
		key := r.Header.Get("X-API-Caller-ID")

		if _, found := getOne(UserCache, key, models.User{}); !found {
			resp.Message = "user nout found: " + key
			resp.Code = http.StatusNotFound

			l.Println(resp.Message, resp.Code)
			break
		}

		if deleted := deleteOne(UserCache, key); !deleted {
			resp.Message = "error deleting: " + key
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			break
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
		break

	case "GET":
		// get user list
		users, _ := getAll(UserCache, models.User{})

		resp.Message = "ok, dumping users"
		resp.Code = http.StatusOK
		resp.Users = users

		l.Println(resp.Message, resp.Code)
		break

	case "POST":
		// post new user
		var user models.User

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			resp.Message = "backend error: cannot read input stream: " + err.Error()
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			break
		}

		data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

		err = json.Unmarshal(data, &user)
		if err != nil {
			resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			break
		}

		if _, found := getOne(UserCache, user.Nickname, models.User{}); found {
			resp.Message = "user already exists"
			resp.Code = http.StatusConflict

			l.Println(resp.Message, resp.Code)
			break
		}

		user.LastActiveTime = time.Now()

		if saved := setOne(UserCache, user.Nickname, user); !saved {
			resp.Message = "backend error: cannot save new user"
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			break
		}

		//resp.Users[user.Nickname] = user

		resp.Message = "ok, adding user"
		resp.Code = http.StatusCreated

		l.Println(resp.Message, resp.Code)
		break

	case "PUT":
		// edit/update an user
		var user models.User

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			resp.Message = "backend error: cannot read input stream: " + err.Error()
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			break
		}

		data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

		err = json.Unmarshal(data, &user)
		if err != nil {
			resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			break
		}

		if _, found := getOne(UserCache, user.Nickname, models.User{}); !found {
			resp.Message = "user not found"
			resp.Code = http.StatusNotFound

			l.Println(resp.Message, resp.Code)
			break
		}

		if saved := setOne(UserCache, user.Nickname, user); !saved {
			resp.Message = "backend error: cannot update the user"
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			break
		}

		resp.Message = "ok, user updated"
		resp.Code = http.StatusCreated

		l.Println(resp.Message, resp.Code)
		break

	default:
		resp.Message = "disallowed method"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		break
	}

	resp.Write(w)
}

func noteUsersActivity(caller string) bool {
	// check if caller exists
	callerUser, found := getOne(UserCache, caller, models.User{})
	if !found {
		return false
	}

	// update user's activity timestamp
	callerUser.LastActiveTime = time.Now()

	return setOne(UserCache, caller, callerUser)
}
