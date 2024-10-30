package db

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/models"
)

// avatarResult is a meta struct to hold the results for various goroutines run in parallel when migrateAvatarURL migration is started.
type avatarResult struct {
	User models.User
	URL  string
}

// defaultAvatarURL is a string returning function to fetch the default URL for the Gravatar service.
var defaultAvatarURL = func() string {
	if os.Getenv("APP_URL_MAIN") != "" {
		return "https://" + os.Getenv("APP_URL_MAIN") + "/web/apple-touch-icon.png"
	}

	return "https://www.littr.eu/web/apple-touch-icon.png"
}()

// GetGravatarURL function returns the avatar image location/URL, or it defaults to a app logo.
func GetGravatarURL(user models.User, channel chan avatarResult, wg *sync.WaitGroup) string {
	// Defer the sync.WaitGroup.Done() when run in goroutine.
	if wg != nil {
		defer wg.Done()
	}

	// Defer the channel closure if it is not nil.
	if channel != nil {
		defer close(channel)
	}

	// A little patch to catch the emailLess accounts naturally (hotfix).
	if user.Email == "" {
		return defaultAvatarURL
	}

	// Preprocess the e-mail address: make it lowercase.
	email := strings.ToLower(user.Email)
	size := 200

	// Prepare the e-mail hash's hexadecimal representation.
	sha := sha256.New()
	sha.Write([]byte(email))

	// Format the hash as a hexadecimal string.
	hashedStringEmail := fmt.Sprintf("%x", sha.Sum(nil))

	// Gravatar base URL.
	url := "https://www.gravatar.com/avatar/" + hashedStringEmail + "?s=" + strconv.Itoa(size) + "&d=" + url.QueryEscape(defaultAvatarURL)

	// Make a GET request towards the Gravatar service.
	resp, err := http.Get(url)
	// On error use the default image URL instead.
	if err != nil {
		url = defaultAvatarImage
	} else {
		defer resp.Body.Close()

		// If the service could not be reached, use the default image.
		if resp.StatusCode != 200 {
			url = defaultAvatarImage
		}
	}

	// Compose the result instance.
	result := avatarResult{
		User: user,
		URL:  url,
	}

	// Write the result to the channel if not nil (already closed for example).
	if channel != nil {
		channel <- result
	}

	return url
}

// fanInChannels is a helper function that collects results from multiple workers.
func fanInChannels(l common.LoggerInterface, channels ...chan avatarResult) <-chan avatarResult {
	var wg sync.WaitGroup

	// Debug log.
	l.Msg(fmt.Sprintf("number of channels to fan-in: %d", len(channels))).Status(http.StatusOK).Log()

	// Common output channel.
	out := make(chan avatarResult)

	// Start a goroutine for each channel to fetch the results.
	for _, channel := range channels {
		wg.Add(1)

		// Assign the goroutine a channel, fetch its result and exit.
		go func(ch <-chan avatarResult) {
			defer wg.Done()
			for result := range ch {
				// Forward the result to the common output channel.
				out <- result
			}
		}(channel)
	}

	// Close the output channel once all worker are done.
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
