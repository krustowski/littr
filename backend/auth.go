package backend

import (
	"litter-go/models"
)

type UserAuth struct {
	User     string `json:"user_name"`
	PassHash string `json:"pass_hash"`
}

func authUser(authUser models.User) (*models.User, bool) {
	users, _ := getAll(UserCache, models.User{})

	for _, user := range users {
		if user.Nickname == authUser.Nickname && user.Passphrase == authUser.Passphrase {
			return &user, true
		}
	}

	return nil, false

	/*user, found := getOne(UserCache, authUser.Nickname, models.User{})
	if !found {
		return nil, false
	}*/

	//return &user, true
}
