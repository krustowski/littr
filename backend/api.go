package backend

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func FlowHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "DELETE":
		// remove a post
		break
	case "GET":
		// get flow
		//GetPosts()
		break
	case "POST":
		// post a new post
		//AddPost()
		break
	case "PUT":
		// edit/update a post
		break
	default:
		response := struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		}{
			Message: "disallowed method",
			Code:    http.StatusBadRequest,
		}
		jsonData, _ := json.Marshal(response)
		io.WriteString(w, fmt.Sprintf("%s", jsonData))
		break
	}
}

func UsersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "DELETE":
		// remove an user
		break
	case "GET":
		// get user list
		break
	case "POST":
		// post new user
		break
	case "PUT":
		// edit/update an user
		break
	default:
		response := struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		}{
			Message: "disallowed method",
			Code:    http.StatusBadRequest,
		}
		jsonData, _ := json.Marshal(response)
		io.WriteString(w, fmt.Sprintf("%s", jsonData))
	}

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
