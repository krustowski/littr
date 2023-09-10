package backend

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"litter-go/models"
)

func AuthHandler(w http.ResponseWriter, r *http.Request) {
	resp := response{}

	lg := models.NewLog(r.RemoteAddr)
	lg.Caller = "[auth] new connection"

	w.Header().Add("Content-Type", "application/json")
	resp.AuthGranted = false

	switch r.Method {
	case "POST":
		var user models.User

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			lg.Content("[auth] " + err.Error())

			resp.Message = "backend error: cannot read input stream"
			resp.Code = http.StatusInternalServerError
			break
		}

		if err = json.Unmarshal(reqBody, &user); err != nil {
			lg.Content("[auth] " + err.Error())

			resp.Message = "backend error: cannot unmarshall request data"
			resp.Code = http.StatusInternalServerError
			break
		}

		// try to authenticate given user
		u, ok := authUser(user)
		if !ok {
			lg.Content("[auth] wrong login")

			resp.Message = "user not found, or wrong passphrase entered"
			resp.Code = http.StatusNotFound
			break
		}

		resp.AuthGranted = ok
		resp.Code = http.StatusOK
		resp.FlowList = u.FlowList
		break

	default:
		resp.Message = "disallowed method"
		resp.Code = http.StatusBadRequest
		break
	}

	resp.Write(w)
}

func FlowHandler(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	log.Println("[flow] new connection from: " + r.RemoteAddr)

	w.Header().Add("Content-Type", "application/json")

	switch r.Method {
	case "DELETE":
		// remove a post
		var post models.Post

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			resp.Message = "backend error: cannot read input stream: " + err.Error()
			resp.Code = http.StatusInternalServerError
			break
		}

		if err := json.Unmarshal(reqBody, &post); err != nil {
			resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
			resp.Code = http.StatusInternalServerError
			break
		}

		key := strconv.FormatInt(post.Timestamp.Unix(), 10)

		if deleted := deleteOne(FlowCache, key); !deleted {
			resp.Message = "cannot delete post"
			resp.Code = http.StatusInternalServerError
			break
		}

		resp.Message = "ok, post removed"
		resp.Code = http.StatusOK
		break

	case "GET":
		// fetch the flow, ergo post list
		posts, _ := getAll(FlowCache, models.Post{})

		resp.Message = "ok, dumping posts"
		resp.Code = http.StatusOK
		resp.Posts = posts
		break

	case "POST":
		// post a new post
		var post models.Post

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			resp.Message = "backend error: cannot read input stream: " + err.Error()
			resp.Code = http.StatusInternalServerError
			break
		}

		if err := json.Unmarshal(reqBody, &post); err != nil {
			resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
			resp.Code = http.StatusInternalServerError
			break
		}

		key := strconv.FormatInt(time.Now().Unix(), 10)

		if saved := setOne(FlowCache, key, post); !saved {
			resp.Message = "backend error: cannot save new post (cache error)"
			resp.Code = http.StatusInternalServerError
			break
		}

		resp.Message = "ok, adding new post"
		resp.Code = http.StatusCreated
		break

	case "PUT":
		// edit/update a post
		var post models.Post

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			resp.Message = "backend error: cannot read input stream: " + err.Error()
			resp.Code = http.StatusInternalServerError
			break
		}

		if err := json.Unmarshal(reqBody, &post); err != nil {
			resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
			resp.Code = http.StatusInternalServerError
			break
		}

		key := strconv.FormatInt(post.Timestamp.Unix(), 10)

		if saved := setOne(FlowCache, key, post); !saved {
			resp.Message = "cannot update post"
			resp.Code = http.StatusInternalServerError
			break
		}

		resp.Message = "ok, post removed"
		resp.Code = http.StatusOK
		break

	default:
		resp.Message = "disallowed method"
		resp.Code = http.StatusBadRequest
		break
	}

	resp.Write(w)
}

func StatsHandler(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	log.Println("[user] new connection from: " + r.RemoteAddr)

	w.Header().Add("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		break

	default:
		resp.Message = "disallowed method"
		resp.Code = http.StatusBadRequest
		break
	}

	resp.Write(w)
}

func UsersHandler(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	log.Println("[user] new connection from: " + r.RemoteAddr)

	w.Header().Add("Content-Type", "application/json")

	switch r.Method {
	case "DELETE":
		// remove an user
		break

	case "GET":
		// get user list
		users, _ := getAll(UserCache, models.User{})

		resp.Message = "ok, dumping users"
		resp.Code = http.StatusOK
		resp.Users = users
		break

	case "POST":
		// post new user
		var user models.User

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			resp.Message = "backend error: cannot read input stream: " + err.Error()
			resp.Code = http.StatusInternalServerError
			break
		}

		err = json.Unmarshal(reqBody, &user)
		if err != nil {
			resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
			resp.Code = http.StatusInternalServerError
			break
		}

		if _, found := getOne(UserCache, user.Nickname, models.User{}); found {
			resp.Message = "user already exists"
			resp.Code = http.StatusConflict
			break
		}

		if saved := setOne(UserCache, user.Nickname, user); !saved {
			resp.Message = "backend error: cannot save new user"
			resp.Code = http.StatusInternalServerError
			break
		}

		//resp.Users[user.Nickname] = user

		resp.Message = "ok, adding user"
		resp.Code = http.StatusCreated
		break

	case "PUT":
		// edit/update an user
		var user models.User

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			resp.Message = "backend error: cannot read input stream: " + err.Error()
			resp.Code = http.StatusInternalServerError
			break
		}

		err = json.Unmarshal(reqBody, &user)
		if err != nil {
			resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
			resp.Code = http.StatusInternalServerError
			break
		}

		if _, found := getOne(UserCache, user.Nickname, models.User{}); !found {
			resp.Message = "user not found"
			resp.Code = http.StatusNotFound
			break
		}

		if saved := setOne(UserCache, user.Nickname, user); !saved {
			resp.Message = "backend error: cannot update the user"
			resp.Code = http.StatusInternalServerError
			break
		}

		//resp.Users[user.Nickname] = user

		resp.Message = "ok, user updated"
		resp.Code = http.StatusCreated
		break

	default:
		resp.Message = "disallowed method"
		resp.Code = http.StatusBadRequest

	}

	resp.Write(w)
}
