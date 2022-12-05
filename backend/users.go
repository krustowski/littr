package backend

import (
	"sort"
	"time"
)

var users []User = []User{
	{Nickname: "krusty", About: "idk lemme just die ffs frfr"},
	{Nickname: "lmao", About: "wtf is this site lmao"},
}

func GetUsers() *[]User {
	// order posts by timestamp DESC
	sort.SliceStable(posts, func(i, j int) bool {
		return posts[i].Timestamp.After(posts[j].Timestamp)
	})

	return &users
}

func AddUser(name, hashedPassphrase, email string) bool {
	var user User = User{
		Nickname:      name,
		Passphrase:    hashedPassphrase,
		Email:         email,
		About:         "new user dropped",
		LastLoginTime: time.Now(),
	}

	users = append(users, user)

	return false
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
