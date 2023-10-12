package backend

import (
	"crypto/md5"
	"encoding/hex"
	"log"
	"net/http"
	"strconv"
	"strings"

	"go.savla.dev/littr/models"
)

func RunMigrations() bool {
	log.Println("migrateAvatarURL():", migrateAvatarURL())
	log.Println("migrateUsersDeletion():", migrateUsersDeletion())

	return true
}

func migrateAvatarURL() bool {
	users, _ := getAll(UserCache, models.User{})

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

func migrateUsersDeletion() bool {
	users, _ := getAll(UserCache, models.User{})
	posts, _ := getAll(FlowCache, models.Post{})

	for key, user := range users {
		if user.Nickname == "fred" || user.Nickname == "fred2" || user.Nickname == "admin" || user.Nickname == "alternative" {
			deleteOne(UserCache, key)
		}
	}

	for key, post := range posts {
		if post.Nickname == "fred" || post.Nickname == "fred2" || post.Nickname == "admin" || post.Nickname == "alternative" {
			deleteOne(FlowCache, key)
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
