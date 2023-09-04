package backend

type UserAuth struct {
	User     string `json:"user_name"`
	PassHash string `json:"pass_hash"`
}

func authUser(authUser User) (*User, bool) {
	/*users, _ := getAll(UserCache, User{})

	for key, user := range users {
		if user.Nickname == authUser.Nickname && user.Passphrase == authUser.Passphrase {
			return &user, true
		}
	}*/

	user, found := getOne(UserCache, authUser.Nickname, User{})
	if !found {
		return nil, false
	}

	return &user, true
}
