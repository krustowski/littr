package backend

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func FlowHandler(w http.ResponseWriter, r *http.Request) {
	response := struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	}{}

	switch r.Method {
	case "DELETE":
		// remove a post
		break
	case "GET":
		// get flow
		break
	case "POST":
		// post a new post
		//AddPost()
		break
	case "PUT":
		// edit/update a post
		break
	default:
		response.Message = "disallowed method"
		response.Code = http.StatusBadRequest

		jsonData, _ := json.Marshal(response)
		io.WriteString(w, fmt.Sprintf("%s", jsonData))
		break
	}
}

func UsersHandler(w http.ResponseWriter, r *http.Request) {
	response := struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
		Users   []User `json:"users"`
	}{}

	switch r.Method {
	case "DELETE":
		// remove an user
		break

	case "GET":
		// get user list
		var users *[]User = GetUsers()
		if users == nil {
			log.Println("error getting user list")
		}

		response.Message = "ok, dumping users"
		response.Code = http.StatusOK
		response.Users = *users

		//io.WriteString(w, fmt.Sprintf("%s", dat))
		break

	case "POST":
		// post new user
		break

	case "PUT":
		// edit/update an user
		break

	default:
		response.Message = "disallowed method"
		response.Code = http.StatusBadRequest

	}

	w.Header().Add("Content-Type", "application/json")
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
