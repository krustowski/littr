package db

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// GetGravatarURL function returns the avatar image location/URL, or it defaults to a app logo.
func GetGravatarURL(emailInput string, channel chan string, wg *sync.WaitGroup) string {
	if wg != nil {
		defer wg.Done()
	}

	email := strings.ToLower(emailInput)
	size := 150

	sha := sha256.New()
	sha.Write([]byte(email))

	hashedStringEmail := fmt.Sprintf("%x", sha.Sum(nil))

	// hash the emailInput
	//byteEmail := []byte(email)
	//hashEmail := md5.Sum(byteEmail)
	//hashedStringEmail := hex.EncodeToString(hashEmail[:])

	url := "https://www.gravatar.com/avatar/" + hashedStringEmail + "?s=" + strconv.Itoa(size)

	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		//log.Println(resp.StatusCode)
		//log.Println(err.Error())
		url = defaultAvatarImage
	} else {
		resp.Body.Close()
	}

	// maybe we are running in a goroutine...
	if channel != nil {
		channel <- url
	}
	return url
}
