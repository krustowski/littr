package backend

import (
	"encoding/json"
	"log"
	"os"
	"sort"
	"time"
)

type users struct {
	Users []User `json:"users"`
}

func GetUsers() *[]User {
	var users users

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

func AddUser(name, hashedPassphrase, email string) bool {
	var users *[]User = GetUsers()
	if users == nil {
		return false
	}

	var user User = User{
		Nickname:      name,
		Passphrase:    hashedPassphrase,
		Email:         email,
		About:         "new user dropped",
		LastLoginTime: time.Now(),
	}

	*users = append(*users, user)

	jsonData, err := json.Marshal(*users)
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

func EditUserPassword(hashedPassphrase string) bool {
	return false
}

func EditUserAbout(aboutText string) bool {
	return false
}

func AuthUser(name, hashedPassword string) bool {
	return false
}

func UserFlowToggle(flowUserName string) bool {
	return false
}
