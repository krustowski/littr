// Application configurations package.
package config

import (
	"fmt"
	"os"
	"strconv"
)

const (
	// Default server port if not specified elsewhere.
	DEFAULT_PORT = "8054"

	// Time interval after that a heartbeat event of type 'message' is to be sent to connected clients/subscribers.
	HEARTBEAT_SLEEP_TIME = 20

	// The anti-duplication variable(s).
	APP_ENVIRONMENT      = "APP_ENVIRONMENT"
	REGISTRATION_ENABLED = "REGISTRATION_ENABLED"
	SERVER_PORT          = "APP_PORT"
)

var (
	// AppEnvironment is a string variable that determines the purpose of the very instance.
	AppEnvironment string = func() string {
		if os.Getenv(APP_ENVIRONMENT) != "" {
			return os.Getenv(APP_ENVIRONMENT)
		} else {
			return "dev"
		}
	}()

	// EnchartedSW is a string variable to hold the templated ServiceWorker contents to load into the very main FE app handler.
	EnchartedSW = func() string {
		// Parse the custom Service Worker template string for the app handler.
		tpl, err := os.ReadFile("/opt/web/app-worker.js.tmpl")
		if err != nil {
			return ""
		}
		return fmt.Sprintf("%s", tpl)
	}()

	// IsLimiterDisabled is a feature flag for the API limiter middleware imported at the APIRouter.
	IsLimiterDisabled bool = func() bool {
		if os.Getenv(LIMITER_DISABLED) == "" {
			return false
		}
		if boolVal, err := strconv.ParseBool(os.Getenv(LIMITER_DISABLED)); err != nil {
			return false
		}
		return boolVal
	}()

	// IsRegistrationEnabled is a boolean to hold the logic for the registration functionality.
	IsRegistrationEnabled bool = func() bool {
		if os.Getenv(REGISTRATION_ENABLED) != "" {
			boolVal, err := strconv.ParseBool(os.Getenv(REGISTRATION_ENABLED))
			if err != nil {
				return false
			}
			return boolVal
		} else {
			return true
		}
	}()

	// ServerPort is a string variable holding the TCP port where the main HTTP server is to listen for incoming connections.
	ServerPort = func() string {
		if os.Getenv(SERVER_PORT) != "" {
			return os.Getenv(SERVER_PORT)
		}

		return DEFAULT_PORT
	}()
)

var (
	// UsersDeletionList holds the list of acounts to be ceased from the database including their posts/polls/assets.
	// The server has to be restarted for changes to apply there. This list also prevents listed nicknames (case insensitive) to be registered.
	UserDeletionList []string = []string{
		"admin",
		"administrator",
		"littr",
		"moderator",
		"nickname",
		"passphrase",
		"superuser",
		"test",
		"tester",
		"user",
		"voter",
	}

	// UsersToUnshade array is used in a procedure's loop to manually unshade listed users.
	// Thus listed accounts should have a zero (0) on stats page.
	UsersToUnshade []string = []string{}
)
