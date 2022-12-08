package backend

import (
	"encoding/json"
	"log"
	"os"
	"sort"
	"time"
)

type Users struct {
	Users []User `json:"users"`
}

func getUsers() *[]User {
	var users Users

	dat, err := os.ReadFile("/opt/data/users.json")
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	err = json.Unmarshal(dat, &users)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	// order posts by timestamp DESC
	sort.SliceStable(users.Users, func(i, j int) bool {
		return users.Users[i].LastActiveTime.After(users.Users[j].LastActiveTime)
	})

	return &users.Users
}

func addUser(user User) bool {
	var users *[]User = getUsers()
	if users == nil {
		return false
	}

	// search for the nickname duplicate
	for _, u := range *users {
		if u.Nickname == user.Nickname {
			log.Println("user already exists!")
			return false
		}
	}

	user.About = "new user dropped"
	user.LastLoginTime = time.Now()

	*users = append(*users, user)

	usersToWrite := &Users{
		Users: *users,
	}

	jsonData, err := json.Marshal(usersToWrite)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	err = os.WriteFile("/opt/data/users.json", jsonData, 0644)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	return true
}

func editUserPassword(hashedPassphrase string) bool {
	return false
}

func editUserAbout(aboutText string) bool {
	return false
}

func userFlowToggle(flowUserName string) bool {
	return false
}
