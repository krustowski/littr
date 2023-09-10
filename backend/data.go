package backend

import (
	"encoding/json"
	"log"
	"os"

	"litter-go/models"
)

const (
	postsFile = "/opt/data/posts.json"
	usersFile = "/opt/data/users.json"
)

var (
	posts = struct{
		Posts map[string]models.Post `json:"posts"`
	}{}

	users = struct{
		Users map[string]models.User `json:"users"`
	}{}
)

func LoadData() {
	rawPostsData, err := os.ReadFile(postsFile)
	if err != nil {
		log.Println(err.Error())
		return
	}

	if string(rawPostsData) == "" {
		return
	}

	err = json.Unmarshal(rawPostsData, &posts)
	if err != nil {
		log.Println(err.Error())
		return
	}

	for key, val := range posts.Posts {
		if key == "" || &val == nil {
			continue
		}

		if saved := setOne(FlowCache, key, val); !saved {
			log.Printf("cannot load post from file (key: %s)", key)
			continue
		}
	}

	rawUsersData, err := os.ReadFile(usersFile)
	if err != nil {
		log.Println(err.Error())
		return
	}

	if string(rawUsersData) == "" {
		return
	}

	err = json.Unmarshal(rawUsersData, &users)
	if err != nil {
		log.Println(err.Error())
		return
	}

	for key, val := range users.Users {
		if key == "" || &val == nil {
			continue
		}

		if saved := setOne(UserCache, key, val); !saved {
			log.Printf("cannot load user from file (key: %s)", key)
			continue
		}
	}
}

func DumpData() {
	posts.Posts, _ = getAll(FlowCache, models.Post{})
	users.Users, _ = getAll(UserCache, models.User{})

	postsJsonData, err := json.Marshal(posts)
	if err != nil {
		log.Println(err.Error())
		return
	}

	usersJsonData, err := json.Marshal(users)
	if err != nil {
		log.Println(err.Error())
		return
	}

	err = os.WriteFile(postsFile, postsJsonData, 0644)
	if err != nil {
		log.Println(err.Error())
		return
	}

	err = os.WriteFile(usersFile, usersJsonData, 0644)
	if err != nil {
		log.Println(err.Error())
		return
	}
}
