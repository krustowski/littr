package backend

import (
	"go.savla.dev/littr/models"
)

func runMigrations() bool {
	return migrateAvatarURL()
}

func migrateAvatarURL() bool {
	users := getAll(UserCache, models.User{})

	for key, user := range users {
		if user.AvatarURL != "" {
			continue
		}

		user.AvatarURL = getGravatarURL(user.Email)
		if ok := setOne(UserCache, key, user); !ok {
			return false
		}
	}

	return true
}

func getGravatarURL(emailInput string) string {
	// TODO: do not hardcode this
	baseURL := "https://littr.n0p.cz/"
	email := strings.ToLower(emailInput)
	size := 150

	defaultImage := "/web/android-chrome-192x192.png"

	byteEmail := []byte(email)
	hashEmail := md5.Sum(byteEmail)
	hashedStringEmail := hex.EncodeToString(hashEmail[:])

	url := "https://www.gravatar.com/avatar/" + hashedStringEmail + "?d=" + baseURL + "&s=" + strconv.Itoa(size)

	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		return defaultImage
	}

	return url
}
