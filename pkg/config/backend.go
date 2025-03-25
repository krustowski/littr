// Application configurations package.
package config

import (
	"os"
	"strconv"
	"time"
)

const (
	// Limiter's settings, limit = req per duration.
	ApiLimiterDuration      time.Duration = 30
	ApiLimiterRequestsCount int           = 100

	// Time interval after that a heartbeat event of type 'message' is to be sent to connected clients/subscribers.
	StreamerHeartbeatPeriod time.Duration = 20
)

const (
	envAppEnvironment      string = "APP_ENVIRONMENT"
	envAppPort             string = "APP_PORT"
	envAppUrl              string = "APP_URL_MAIN"
	envAppVersion          string = "APP_VERSION"
	envDataDumpFormat      string = "DATA_DUMP_FORMAT"
	envDataLoadFormat      string = "DATA_LOAD_FORMAT"
	envDockerInternalPort  string = "DOCKER_INTERNAL_PORT"
	envDumpToken           string = "APP_TOKEN"
	envLimiterEnabled      string = "LIMITER_ENABLED"
	envRegistrationEnabled string = "REGISTRATION_ENABLED"
	envServerSecret        string = "APP_PEPPER"
	envServerPort          string = "SERVER_PORT"
)

const serviceWorkerTemplateFile string = "/opt/web/app-worker.js.tmpl"

const (
	defaultApiLimiterEnabled     bool   = true
	defaultAppEnvironment        string = "dev"
	defaultAppUrl                string = "https://www.littr.eu"
	defaultDataDumpFormat        string = "JSON"
	defaultDataLoadFormat        string = "JSON"
	defaultDumpToken             string = ""
	defaultRegistrationEnabled   bool   = true
	defaultServerPort            string = "8054"
	defaultServerSecret          string = ""
	defaultServiceWorkerTemplate string = ""
	defaultTestServerPort        string = "8777"
	DefaultTestStreamerPort      string = "8778"
)

var (
	// AppEnvironment is a string variable that determines the purpose of the very instance.
	AppEnvironment string = func() string {
		if val := os.Getenv(envAppEnvironment); val != "" {
			return val
		}

		return defaultAppEnvironment
	}()

	AppVersion string = func() string {
		if val := os.Getenv(envAppVersion); val != "" {
			return val
		}

		return "unknown"
	}()

	DataDumpFormat string = func() string {
		if val := os.Getenv(envDataDumpFormat); val != "" {
			return val
		}

		return defaultDataDumpFormat
	}()

	DataDumpToken string = func() string {
		if val := os.Getenv(envDumpToken); val != "" {
			return val
		}

		return defaultDumpToken
	}()

	DataLoadFormat string = func() string {
		if val := os.Getenv(envDataLoadFormat); val != "" {
			return val
		}

		return defaultDataLoadFormat
	}()

	// EnchartedSW is a string variable to hold the templated ServiceWorker contents to load into the very main FE app handler.
	EnchartedSW = func() string {
		// Parse the custom Service Worker template string for the app handler.
		tpl, err := os.ReadFile(serviceWorkerTemplateFile)
		if err != nil {
			return defaultServiceWorkerTemplate
		}

		return string(tpl)
	}()

	// IsApiLimiterEnabled is a feature flag for the API limiter middleware imported at the APIRouter.
	IsApiLimiterEnabled bool = func() bool {
		if val := os.Getenv(envLimiterEnabled); val != "" {
			boolVal, err := strconv.ParseBool(val)
			if err != nil {
				return false
			}

			return boolVal
		}

		return defaultApiLimiterEnabled
	}()

	// IsRegistrationEnabled is a boolean to hold the logic for the registration functionality.
	IsRegistrationEnabled bool = func() bool {
		if val := os.Getenv(envRegistrationEnabled); val != "" {
			boolVal, err := strconv.ParseBool(val)
			if err != nil {
				return false
			}

			return boolVal
		}

		return defaultRegistrationEnabled
	}()

	// ServerPort is a string variable holding the TCP port where the main HTTP server is to listen for incoming connections.
	ServerPort = func() string {
		if val := os.Getenv(envServerPort); val != "" {
			return val
		}

		// Try the alternative(s) if SERVER_PORT is blank.
		if val := os.Getenv(envDockerInternalPort); val != "" {
			return val
		}

		if val := os.Getenv(envAppPort); val != "" {
			return val
		}

		return defaultServerPort
	}()

	ServerSecret = func() string {
		if val := os.Getenv(envServerSecret); val != "" {
			return val
		}

		return defaultServerSecret
	}()

	ServerUrl = func() string {
		if val := os.Getenv(envAppUrl); val != "" {
			return val
		}

		return defaultAppUrl
	}()
)

var (
	// UsersDeletionList holds the list of acounts to be ceased from the database including their posts/polls/assets.
	// The server has to be restarted for changes to apply there. This list also prevents listed nicknames (case insensitive) to be registered.
	UserDeletionList []string = []string{
		"activation",
		"admin",
		"administrator",
		"caller",
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
