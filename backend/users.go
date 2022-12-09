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

// https://stackoverflow.com/a/37335777
func remove(slice []string, s int) []string {
	return append(slice[:s], slice[s+1:]...)
}

func userFlowToggle(user User) (*User, bool) {
	usersP := getUsers()
	if usersP == nil {
		return nil, false
	}
	users := *usersP

	var i int
	var u User
	for idx, u := range users {
		// found user
		if u.Nickname == user.Nickname {
			i = idx
			break
		}

		i = -1
	}

	if i < 0 {
		return nil, false
	}

	var j int
	var recordFound bool = false
	for idx, rec := range u.Flow {
		if rec == user.FlowToggle {
			recordFound = true
			j = idx
			break
		}
	}

	// remove from flow list
	if recordFound {
		user.Flow = remove(user.Flow, j)
	} else {
		user.Flow = append(user.Flow, user.FlowToggle)
	}

	u.Flow = user.Flow
	users[i] = u

	usersToWrite := &Users{
		Users: users,
	}

	jsonData, err := json.Marshal(usersToWrite)
	if err != nil {
		log.Println(err.Error())
		return nil, false
	}

	err = os.WriteFile("/opt/data/users.json", jsonData, 0644)
	if err != nil {
		log.Println(err.Error())
		return nil, false
	}

	return &user, true
}
