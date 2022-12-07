package backend

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

func FlowHandler(w http.ResponseWriter, r *http.Request) {
	response := struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
		Posts   []Post `json:"posts"`
	}{}
	w.Header().Add("Content-Type", "application/json")

	switch r.Method {
	case "DELETE":
		// remove a post
		break
	case "GET":
		// get flow, ergo post list
		var posts *[]Post = GetPosts()
		if posts == nil {
			log.Println("error getting post flow list")
			return
		}

		response.Message = "ok, dumping posts"
		response.Code = http.StatusOK
		response.Posts = *posts
		break
	case "POST":
		// post a new post
		var post Post

		reqBody, _ := ioutil.ReadAll(r.Body)
		err := json.Unmarshal(reqBody, &post)
		if err != nil {
			log.Println(err.Error())
			return
		}

		if ok := AddPost(post.Content); !ok {
			log.Println("error adding new post")
			return
		}

		response.Message = "ok, adding post"
		response.Code = http.StatusCreated
		response.Posts = append(response.Posts, post)
		break
	case "PUT":
		// edit/update a post
		break
	default:
		response.Message = "disallowed method"
		response.Code = http.StatusBadRequest

		break
	}

	jsonData, _ := json.Marshal(response)
	io.WriteString(w, fmt.Sprintf("%s", jsonData))
}

func UsersHandler(w http.ResponseWriter, r *http.Request) {
	response := struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
		Users   []User `json:"users"`
	}{}
	w.Header().Add("Content-Type", "application/json")

	switch r.Method {
	case "DELETE":
		// remove an user
		break

	case "GET":
		// get user list
		var users *[]User = GetUsers()
		if users == nil {
			log.Println("error getting user list")
			return
		}

		response.Message = "ok, dumping users"
		response.Code = http.StatusOK
		response.Users = *users
		break

	case "POST":
		// post new user
		var user User

		reqBody, _ := ioutil.ReadAll(r.Body)
		err := json.Unmarshal(reqBody, &user)
		if err != nil {
			log.Println(err.Error())
			return
		}

		if ok := AddUser(user); !ok {
			log.Println("error adding new user")
			return
		}

		response.Message = "ok, adding user"
		response.Code = http.StatusCreated
		response.Users = append(response.Users, user)
		break

	case "PUT":
		// edit/update an user
		break

	default:
		response.Message = "disallowed method"
		response.Code = http.StatusBadRequest

	}

	jsonData, _ := json.Marshal(response)
	io.WriteString(w, fmt.Sprintf("%s", jsonData))

	// process list requests
	if hasList := r.URL.Query().Has("list"); hasList {
		list := r.URL.Query().Get("list")

		switch list {
		case "flow":
			io.WriteString(w, fmt.Sprintf("flow -- %s\n", list))
			break
		case "users":

			io.WriteString(w, fmt.Sprintf("users -- %s\n", list))
			break
		}

		/*body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Printf("could not read body: %s\n", err)
		}*/
	}

	newPost := r.PostFormValue("newPost")
	if newPost != "" {
		io.WriteString(w, fmt.Sprintf("POST -- newPost"))
	}
}
