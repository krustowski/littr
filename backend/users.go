package backend

import (
	"log"
	"net/mail"
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
		return posts[i].Timestamp > posts[j].Timestamp
	})

	return &users
}

func AddUser(name, hashedPassphrase, email string) bool {
	// validate e-mail struct
	// https://stackoverflow.com/a/66624104
	if _, err := mail.ParseAddress(email); err != nil {
		log.Println(err)
		return false
	}

	var user User = User{
		Nickname:      name,
		Passphrase:    hashedPassphrase,
		Email:         email,
		About:         "new user dropped",
		LastLoginTime: time.Now(),
	}

	users = append(users, user)

	return true
}

func EditUserPassword(hashedPassphrase string) bool {
	return true
}
