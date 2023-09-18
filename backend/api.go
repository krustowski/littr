package backend

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.savla.dev/littr/config"
	"go.savla.dev/littr/models"
)

func AuthHandler(w http.ResponseWriter, r *http.Request) {
	resp := response{}

	//lg := models.NewLog(r.RemoteAddr)
	//lg.Caller = "[auth] new connection"
	log.Println("[auth] new connection from: " + r.RemoteAddr)

	resp.AuthGranted = false

	switch r.Method {
	case "POST":
		var user models.User

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("[auth] " + err.Error())

			resp.Message = "backend error: cannot read input stream"
			resp.Code = http.StatusInternalServerError
			break
		}

		data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

		if err = json.Unmarshal(data, &user); err != nil {
			log.Println("[auth] " + err.Error())

			resp.Message = "backend error: cannot unmarshall request data"
			resp.Code = http.StatusInternalServerError
			break
		}

		// try to authenticate given user
		u, ok := authUser(user)
		if !ok {
			log.Println("[auth] wrong login")

			resp.Message = "user not found, or wrong passphrase entered"
			resp.Code = http.StatusBadRequest
			break
		}

		resp.Users = make(map[string]models.User)
		resp.Users[u.Nickname] = *u
		resp.AuthGranted = ok
		resp.Code = http.StatusOK
		//resp.FlowList = u.FlowList
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

	switch r.Method {
	case "DELETE":
		// remove a post
		var post models.Post

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("backend error: cannot read input stream: " + err.Error())
			resp.Message = "backend error: cannot read input stream: " + err.Error()
			resp.Code = http.StatusInternalServerError
			break
		}

		data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

		if err := json.Unmarshal(data, &post); err != nil {
			log.Println("backend error: cannot unmarshall fetched data: " + err.Error())
			resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
			resp.Code = http.StatusInternalServerError
			break
		}

		//key := strconv.FormatInt(post.Timestamp.UnixNano(), 10)
		key := post.ID

		if deleted := deleteOne(FlowCache, key); !deleted {
			log.Println("cannot delete the post")
			resp.Message = "cannot delete the post"
			resp.Code = http.StatusInternalServerError
			break
		}

		log.Println("ok, post removed")
		resp.Message = "ok, post removed"
		resp.Code = http.StatusOK
		break

	case "GET":
		// fetch the flow, ergo post list
		posts, count := getAll(FlowCache, models.Post{})

		resp.Message = "ok, dumping posts"
		resp.Code = http.StatusOK
		resp.Posts = posts
		resp.Count = count
		break

	case "POST":
		// post a new post
		var post models.Post

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("backend error: cannot read input stream: " + err.Error())
			resp.Message = "backend error: cannot read input stream: " + err.Error()
			resp.Code = http.StatusInternalServerError
			break
		}

		data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

		if err := json.Unmarshal(data, &post); err != nil {
			log.Println("backend error: cannot unmarshall fetched data: " + err.Error())
			resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
			resp.Code = http.StatusInternalServerError
			break
		}

		key := strconv.FormatInt(time.Now().UnixNano(), 10)
		post.ID = key

		if saved := setOne(FlowCache, key, post); !saved {
			log.Println("backend error: cannot save new post (cache error)")
			resp.Message = "backend error: cannot save new post (cache error)"
			resp.Code = http.StatusInternalServerError
			break
		}

		log.Println("ok, adding new post")
		resp.Message = "ok, adding new post"
		resp.Code = http.StatusCreated
		break

	case "PUT":
		// edit/update a post
		var post models.Post

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("backend error: cannot read input stream: " + err.Error())
			resp.Message = "backend error: cannot read input stream: " + err.Error()
			resp.Code = http.StatusInternalServerError
			break
		}

		data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

		if err := json.Unmarshal(data, &post); err != nil {
			log.Println("backend error: cannot unmarshall fetched data: " + err.Error())
			resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
			resp.Code = http.StatusInternalServerError
			break
		}

		//key := strconv.FormatInt(post.Timestamp.UnixNano(), 10)
		key := post.ID

		/*if _, found := getOne(FlowCache, key, models.User{}); !found {
			log.Println("unknown post update requested")
			resp.Message = "unknown post update requested"
			resp.Code = http.StatusBadRequest
			break
		}*/

		if saved := setOne(FlowCache, key, post); !saved {
			resp.Message = "cannot update post"
			resp.Code = http.StatusInternalServerError
			break
		}

		log.Println("ok, post updated")
		resp.Message = "ok, post updated"
		resp.Code = http.StatusOK
		break

	default:
		resp.Message = "disallowed method"
		resp.Code = http.StatusBadRequest
		break
	}

	resp.Write(w)
}

func PollsHandler(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	log.Println("[poll] new connection from: " + r.RemoteAddr)

	switch r.Method {
	case "DELETE":
		var poll models.Poll

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("backend error: cannot read input stream: " + err.Error())
			resp.Message = "backend error: cannot read input stream: " + err.Error()
			resp.Code = http.StatusInternalServerError
			break
		}

		data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

		if err := json.Unmarshal(data, &poll); err != nil {
			log.Println("backend error: cannot unmarshall fetched data: " + err.Error())
			resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
			resp.Code = http.StatusInternalServerError
			break
		}

		key := poll.ID

		if deleted := deleteOne(PollCache, key); !deleted {
			log.Println("cannot delete the poll")
			resp.Message = "cannot delete the poll"
			resp.Code = http.StatusInternalServerError
			break
		}

		log.Println("ok, poll removed")
		resp.Message = "ok, poll removed"
		resp.Code = http.StatusOK
		break

	case "GET":
		polls, _ := getAll(PollCache, models.Poll{})

		resp.Message = "ok, dumping polls"
		resp.Code = http.StatusOK
		resp.Polls = polls
		break

	case "POST":
		var poll models.Poll

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("backend error: cannot read input stream: " + err.Error())
			resp.Message = "backend error: cannot read input stream: " + err.Error()
			resp.Code = http.StatusInternalServerError
			break
		}

		data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

		if err := json.Unmarshal(data, &poll); err != nil {
			log.Println("backend error: cannot unmarshall fetched data: " + err.Error())
			resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
			resp.Code = http.StatusInternalServerError
			break
		}

		key := poll.ID

		if saved := setOne(PollCache, key, poll); !saved {
			log.Println("backend error: cannot save new poll (cache error)")
			resp.Message = "backend error: cannot save new poll (cache error)"
			resp.Code = http.StatusInternalServerError
			break
		}

		log.Println("ok, adding new poll")
		resp.Message = "ok, adding new poll"
		resp.Code = http.StatusCreated
		break

	case "PUT":
		var poll models.Poll

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("backend error: cannot read input stream: " + err.Error())
			resp.Message = "backend error: cannot read input stream: " + err.Error()
			resp.Code = http.StatusInternalServerError
			break
		}

		data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

		if err := json.Unmarshal(data, &poll); err != nil {
			log.Println("backend error: cannot unmarshall fetched data: " + err.Error())
			resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
			resp.Code = http.StatusInternalServerError
			break
		}

		//key := strconv.FormatInt(post.Timestamp.UnixNano(), 10)
		key := poll.ID

		if _, found := getOne(PollCache, key, models.Poll{}); !found {
			log.Println("unknown poll update requested")
			resp.Message = "unknown poll update requested"
			resp.Code = http.StatusBadRequest
			break
		}

		if saved := setOne(PollCache, key, poll); !saved {
			resp.Message = "cannot update post"
			resp.Code = http.StatusInternalServerError
			break
		}

		log.Println("ok, poll updated")
		resp.Message = "ok, poll updated"
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
	log.Println("[stats] new connection from: " + r.RemoteAddr)

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

		data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

		err = json.Unmarshal(data, &user)
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

		data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

		err = json.Unmarshal(data, &user)
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
