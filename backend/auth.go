package backend

type Auth struct{}

func authUser(user User) (*User, bool) {
	var users *[]User = getUsers()

	for _, u := range *users {
		if u.Nickname == user.Nickname && u.Passphrase == user.Passphrase {
			return &u, true
		}
	}

	return nil, false
}
