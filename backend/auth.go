package backend

import (
	"go.savla.dev/littr/models"
)

type UserAuth struct {
	User     string `json:"user_name"`
	PassHash string `json:"pass_hash"`
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	resp := response{}

	// prepare the Logger instance
	l := Logger{
		IPAddress:  r.Header.Get("X-Real-IP"),
		Method:     r.Method,
		Route:      r.URL.String(),
		WorkerName: "auth",
		Version:    r.Header.Get("X-App-Version"),
	}

	resp.AuthGranted = false

	var user models.User

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		resp.Message = "backend error: cannot read input stream"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message+err.Error(), resp.Code)
		break
	}

	data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

	if err = json.Unmarshal(data, &user); err != nil {
		resp.Message = "backend error: cannot unmarshall request data"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message+err.Error(), resp.Code)
		break
	}

	l.CallerID = user.Nickname

	// try to authenticate given user
	u, ok := authUser(user)
	if !ok {
		resp.Message = "user not found, or wrong passphrase entered"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		break
	}

	resp.Users = make(map[string]models.User)
	resp.Users[u.Nickname] = *u
	resp.AuthGranted = ok

	resp.Message = "auth granted"
	resp.Code = http.StatusOK
	//resp.FlowList = u.FlowListdd

	l.Println(resp.Message, resp.Code)

	resp.Write(w)
}

func authUser(aUser models.User) (*models.User, bool) {
	// fetch one user from cache according to the login credential
	user, ok := getOne(UserCache, aUser.Nickname, models.User{})
	if !ok {
		// not found
		return nil, false
	}

	// check the passhash
	if user.Passphrase == aUser.Passphrase {
		// auth granted
		return &user, true
	}

	// auth failed
	return nil, false
}
