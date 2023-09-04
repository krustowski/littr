package backend

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type response struct {
	AuthGranted bool            `json:"auth_granted" default:false`
	Code        int             `json:"code"`
	FlowList    []string        `json:"flow_records"`
	Key         string          `json:"key"`
	Message     string          `json:"message"`
	Posts       map[string]Post `json:"posts"`
	Users       map[string]User `json:"users"`
}

func AuthHandler(w http.ResponseWriter, r *http.Request) {
	resp := response{}

	w.Header().Add("Content-Type", "application/json")

	switch r.Method {
	case "POST":
		var user User

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			resp.Message = "backend error: cannot read input stream: " + err.Error()
			resp.Code = http.StatusInternalServerError
			break
		}

		if err = json.Unmarshal(reqBody, &user); err != nil {
			resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
			resp.Code = http.StatusInternalServerError
			break
		}

		// try to authenticate given user
		u, ok := authUser(user)
		if !ok {
			resp.Message = "user not found, or wrong passphrase entered"
			resp.Code = http.StatusNotFound
			resp.AuthGranted = false
			return
		}

		resp.AuthGranted = true
		resp.FlowList = u.FlowList
		break
	default:
		resp.Message = "disallowed method"
		resp.Code = http.StatusBadRequest
		break
	}

	jsonData, _ := json.Marshal(resp)
	io.WriteString(w, fmt.Sprintf("%s", jsonData))
}

func FlowHandler(w http.ResponseWriter, r *http.Request) {
	resp := response{}

	w.Header().Add("Content-Type", "application/json")

	switch r.Method {
	case "DELETE":
		// remove a post
		break
	case "GET":
		// fetch the flow, ergo post list
		posts, _ := getAll(FlowCache, Post{})

		resp.Message = "ok, dumping posts"
		resp.Code = http.StatusOK
		resp.Posts = posts
		break
	case "POST":
		// post a new post
		var post Post

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

		resp.Posts[key] = post

		resp.Message = "ok, adding new post"
		resp.Code = http.StatusCreated
		break
	case "PUT":
		// edit/update a post
		break
	default:
		resp.Message = "disallowed method"
		resp.Code = http.StatusBadRequest
		break
	}

	jsonData, _ := json.Marshal(resp)
	io.WriteString(w, fmt.Sprintf("%s", jsonData))
}

func UsersHandler(w http.ResponseWriter, r *http.Request) {
	resp := response{}

	//r.Header.Get("X-System-Token")
	w.Header().Add("Content-Type", "application/json")

	switch r.Method {
	case "DELETE":
		// remove an user
		break

	case "GET":
		// get user list
		users, _ := getAll(UserCache, User{})

		resp.Message = "ok, dumping users"
		resp.Code = http.StatusOK
		resp.Users = users
		break

	case "POST":
		// post new user
		var user User

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

		if _, found := getOne(UserCache, user.Nickname, User{}); found {
			resp.Message = "user already exists"
			resp.Code = http.StatusConflict
			break
		}

		if saved := setOne(UserCache, user.Nickname, user); !saved {
			resp.Message = "backend error: cannot save new user"
			resp.Code = http.StatusInternalServerError
			break
		}

		resp.Users[user.Nickname] = user

		resp.Message = "ok, adding user"
		resp.Code = http.StatusCreated
		break

	case "PUT":
		// edit/update an user
		var user User

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

		if _, found := getOne(UserCache, user.Nickname, User{}); !found {
			resp.Message = "user not found"
			resp.Code = http.StatusNotFound
			break
		}

		if saved := setOne(UserCache, user.Nickname, user); !saved {
			resp.Message = "backend error: cannot update the user"
			resp.Code = http.StatusInternalServerError
			break
		}

		resp.Users[user.Nickname] = user

		resp.Message = "ok, user updated"
		resp.Code = http.StatusCreated
		break

	default:
		resp.Message = "disallowed method"
		resp.Code = http.StatusBadRequest

	}

	// send JSON response
	jsonData, _ := json.Marshal(resp)
	io.WriteString(w, fmt.Sprintf("%s", jsonData))
}
