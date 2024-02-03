package backend

import (
	"io"
	"net/http"
	"os"

	"go.savla.dev/littr/config"
	"go.savla.dev/littr/models"
)

func getUsers(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := Logger{
		CallerID:   r.Header.Get("X-API-Caller-ID"),
		IPAddress:  r.Header.Get("X-Real-IP"),
		Method:     r.Method,
		Route:      r.URL.String(),
		WorkerName: "users",
		Version:    r.Header.Get("X-App-Version"),
	}
	noteUsersActivity(r.Header.Get("X-API-Caller-ID"))

	// get user list
	users, _ := getAll(UserCache, models.User{})

	resp.Message = "ok, dumping users"
	resp.Code = http.StatusOK
	resp.Users = users

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
}

func addNewUser(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := Logger{
		CallerID:   r.Header.Get("X-API-Caller-ID"),
		IPAddress:  r.Header.Get("X-Real-IP"),
		Method:     r.Method,
		Route:      r.URL.String(),
		WorkerName: "users",
		Version:    r.Header.Get("X-App-Version"),
	}
	noteUsersActivity(r.Header.Get("X-API-Caller-ID"))

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
	resp.Write(w)
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := Logger{
		CallerID:   r.Header.Get("X-API-Caller-ID"),
		IPAddress:  r.Header.Get("X-Real-IP"),
		Method:     r.Method,
		Route:      r.URL.String(),
		WorkerName: "users",
		Version:    r.Header.Get("X-App-Version"),
	}
	noteUsersActivity(r.Header.Get("X-API-Caller-ID"))

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
	resp.Write(w)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := Logger{
		CallerID:   r.Header.Get("X-API-Caller-ID"),
		IPAddress:  r.Header.Get("X-Real-IP"),
		Method:     r.Method,
		Route:      r.URL.String(),
		WorkerName: "users",
		Version:    r.Header.Get("X-App-Version"),
	}
	noteUsersActivity(r.Header.Get("X-API-Caller-ID"))

	key := r.Header.Get("X-API-Caller-ID")

	if _, found := getOne(UserCache, key, models.User{}); !found {
		resp.Message = "user nout found: " + key
		resp.Code = http.StatusNotFound

		l.Println(resp.Message, resp.Code)
		break
	}

	if deleted := deleteOne(UserCache, key); !deleted {
		resp.Message = "error deleting:" + key
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
	resp.Write(w)
}
