package backend

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

func AuthHandler(w http.ResponseWriter, r *http.Request) {
	response := struct {
		Message     string   `json:"message"`
		Code        int      `json:"code"`
		AuthGranted bool     `json:"auth_granted" default:false`
		FlowRecords []string `json:"flow_records"`
	}{}
	w.Header().Add("Content-Type", "application/json")

	switch r.Method {
	case "POST":
		var user User

		reqBody, _ := ioutil.ReadAll(r.Body)
		err := json.Unmarshal(reqBody, &user)
		if err != nil {
			log.Println(err.Error())
			response.Message = err.Error()
			return
		}

		u, ok := authUser(user)
		if !ok {
			response.Message = "user not found or wrong passphrase entered"
			response.Code = http.StatusNotFound
			response.AuthGranted = false
			return
		}

		response.AuthGranted = true
		response.FlowRecords = u.Flow
		break
	default:
		response.Message = "disallowed method"
		response.Code = http.StatusBadRequest
		break
	}

	jsonData, _ := json.Marshal(response)
	io.WriteString(w, fmt.Sprintf("%s", jsonData))
}

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
		var posts *[]Post = getPosts()
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

		if ok := addPost(post); !ok {
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
		Message     string   `json:"message"`
		Code        int      `json:"code"`
		Users       []User   `json:"users"`
		FlowRecords []string `json:"flow_records"`
	}{}
	w.Header().Add("Content-Type", "application/json")

	switch r.Method {
	case "DELETE":
		// remove an user
		break

	case "GET":
		// get user list
		var users *[]User = getUsers()
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

		if ok := addUser(user); !ok {
			log.Println("error adding new user")
			return
		}

		response.Message = "ok, adding user"
		response.Code = http.StatusCreated
		response.Users = append(response.Users, user)
		break

	case "PUT":
		// edit/update an user
		var reqUser User

		reqBody, _ := ioutil.ReadAll(r.Body)
		err := json.Unmarshal(reqBody, &reqUser)
		if err != nil {
			log.Println(err.Error())
			return
		}

		// userFlowToggle(User)
		if reqUser.FlowToggle != "" {
			u, ok := userFlowToggle(reqUser)
			if !ok {
				log.Println("error updating user's flow")
				return
			}
			response.FlowRecords = u.Flow
		}

		if reqUser.About != "" {
			_, ok := editUserAbout(reqUser)
			if !ok {
				log.Println("error updating user's about text")
				return
			}
		}

		if reqUser.Passphrase != "" {
			_, ok := editUserPassphrase(reqUser)
			if !ok {
				log.Println("error updating user's passphrase")
				return
			}
		}

		//response.Message = "ok, adding user"
		response.Code = http.StatusOK
		break

	default:
		response.Message = "disallowed method"
		response.Code = http.StatusBadRequest

	}

	jsonData, _ := json.Marshal(response)
	io.WriteString(w, fmt.Sprintf("%s", jsonData))
}
