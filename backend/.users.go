package backend

import (
	"encoding/json"
	"log"
	"os"
	"sort"
	"time"
)

const (
	usersFile = "/opt/data/users.json"
)

type Users struct {
	Users []User `json:"users"`
}

func findUser(users []User, user User) (int, *User) {
	for idx, u := range users {
		if u.Nickname == user.Nickname {
			return idx, &u
		}
	}
	return -1, nil
}

func writeUsers(users *[]User) bool {
	usersToWrite := &Users{
		Users: *users,
	}

	jsonData, err := json.Marshal(usersToWrite)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	err = os.WriteFile(usersFile, jsonData, 0644)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	return true
}

func getUsers() *[]User {
	var users Users

	dat, err := os.ReadFile(usersFile)
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
	if _, foundUser := findUser(*users, user); foundUser != nil {
		log.Println("user already exists!")
		return false
	}

	user.About = "new user dropped"
	user.LastLoginTime = time.Now()

	*users = append(*users, user)

	if ok := writeUsers(users); !ok {
		return false
	}

	return true
}

func editUserPassphrase(user User) (*User, bool) {
	usersP := getUsers()
	if usersP == nil {
		return nil, false
	}
	users := *usersP

	idx, foundUser := findUser(users, user)
	if foundUser == nil {
		return nil, false
	}

	// rewrite user's passphrase
	foundUser.Passphrase = user.Passphrase

	users[idx] = *foundUser

	if ok := writeUsers(&users); !ok {
		log.Println("cannot write users")
		return nil, false
	}

	return foundUser, true
}

func editUserAbout(user User) (*User, bool) {
	usersP := getUsers()
	if usersP == nil {
		return nil, false
	}
	users := *usersP

	idx, foundUser := findUser(users, user)
	if foundUser == nil {
		return nil, false
	}

	// rewrite user's about text
	foundUser.About = user.About

	users[idx] = *foundUser

	if ok := writeUsers(&users); !ok {
		log.Println("cannot write users")
		return nil, false
	}

	return foundUser, true
}

func remove(items []string, idx int) []string {
	// delete an element from the array
	// https://www.educative.io/answers/how-to-delete-an-element-from-an-array-in-golang
	newLength := 0
	for index := range items {
		if idx != index {
			items[newLength] = items[index]
			newLength++
		}
	}
	items = items[:newLength]

	return items
}

func userFlowToggle(user User) (*User, bool) {
	usersP := getUsers()
	if usersP == nil {
		return nil, false
	}
	users := *usersP

	// check for existing user
	i, foundUser := findUser(users, user)
	if foundUser == nil {
		return nil, false
	}

	// find flow record
	var j int
	var recordFound bool = false
	for idx, rec := range foundUser.Flow {
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

	// copy modified (user) flow array to the production (foundUser) struct
	users[i].Flow = user.Flow

	if ok := writeUsers(&users); !ok {
		return nil, false
	}

	return &user, true
}
