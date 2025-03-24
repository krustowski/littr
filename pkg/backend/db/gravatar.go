package db

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/models"
)

var (
	ErrHttpRequestCreationFailed = errors.New("could not create new HTTP request")
	ErrHttpRequestUsageFailed    = errors.New("could not use the HTTP request")
	ErrResponseBodyMisbehave     = errors.New("could not close the response body: ")
)

// avatarResult is a meta struct to hold the results for the avatar migration's (migrateAvatarURL) channels.
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
func GetGravatarURL(user models.User, channel chan interface{}, wg *sync.WaitGroup, client *http.Client) string {
	l := common.NewLogger(nil, "gravatar")

	var result avatarResult

	defer func() {
		if channel != nil {
			channel <- &result
			close(channel)
		}

		if wg != nil {
			wg.Done()
		}
	}()

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

	// Create new contexted request.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		l.Msg(ErrHttpRequestCreationFailed.Error()).Error(err).Status(http.StatusInternalServerError).Log()

		result = avatarResult{
			User: user,
			URL:  defaultAvatarURL,
		}

		return defaultAvatarURL
	}

	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	// Make a GET request towards the Gravatar service.
	resp, err := client.Do(req)
	// On error use the default image URL instead.
	if err != nil {
		l.Msg(ErrHttpRequestUsageFailed.Error()).Error(err).Status(http.StatusInternalServerError).Log()

		result = avatarResult{
			User: user,
			URL:  defaultAvatarURL,
		}

		return defaultAvatarURL
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			l.Msg(ErrResponseBodyMisbehave.Error()).Error(err).Log()
		}
	}()

	// If the service could not be reached, use the default image.
	if resp.StatusCode != 200 {
		url = defaultAvatarURL
	}

	// Compose the result instance.
	result = avatarResult{
		User: user,
		URL:  url,
	}

	return url
}

// FanInChannels is a helper function that collects results from multiple workers.
func FanInChannels(l common.Logger, channels ...chan interface{}) <-chan interface{} {
	var wg sync.WaitGroup

	// Debug log.
	if l != nil {
		l.Msg(fmt.Sprintf("number of channels to fan-in: %d", len(channels))).Status(http.StatusOK).Log()
	}

	// Common output channel.
	out := make(chan interface{}, 1)

	// Start a goroutine for each channel to fetch the results.
	for _, channel := range channels {
		wg.Add(1)

		// Assign the goroutine a channel, fetch its result and exit.
		go func(ch <-chan interface{}) {
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
