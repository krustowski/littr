// @title		littr
// @version	 	0.44.34
// @description		a simple nanoblogging platform as PWA built on go-app framework
// @termsOfService	https://www.littr.eu/tos

// @contact.name	API Support
// @contact.url		https://www.littr.eu/docs
// @contact.email	info@littr.eu

// @license.name	MIT
// @license.url		https://github.com/krustowski/littr/blob/master/LICENSE

// @host		www.littr.eu
// @BasePath		/api/v1
// @accept              json
// @produce             json

// @supportedSubmitMethods	[]
// @security            	[]
// @securityDefinitions.basic	BasicAuth

// @externalDocs.description	OpenAPI
// @externalDocs.url		https://swagger.io/resources/open-api/

// @externalDocs.description	Documentation
// @externalDocs.url		https://krusty.space/projects/littr/

// The umbrella package for the JSON REST API service.
package backend

import (
	"net/http"
	"os"
	"time"

	//"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"

	"go.vxn.dev/littr/pkg/backend/auth"
	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/backend/db"
	"go.vxn.dev/littr/pkg/backend/live"
	"go.vxn.dev/littr/pkg/backend/polls"
	"go.vxn.dev/littr/pkg/backend/posts"
	"go.vxn.dev/littr/pkg/backend/push"
	"go.vxn.dev/littr/pkg/backend/requests"
	"go.vxn.dev/littr/pkg/backend/stats"
	"go.vxn.dev/littr/pkg/backend/tokens"
	"go.vxn.dev/littr/pkg/backend/users"
	"go.vxn.dev/littr/pkg/config"
)

var (
	// Default HTTP 404 response.
	NotFoundHandler = func(w http.ResponseWriter, r *http.Request) {
		dummyRootLoggerWriter(w, r, "page not found", http.StatusNotFound)
	}

	// Default HTTP 405 response.
	MethodNotAllowedHandler = func(w http.ResponseWriter, r *http.Request) {
		dummyRootLoggerWriter(w, r, "method not allowed", http.StatusMethodNotAllowed)
	}

	// The JSON API service root path handler (served at /api/v1).
	rootHandler = func(w http.ResponseWriter, r *http.Request) {
		msg := "littr JSON API service (v" + os.Getenv("APP_VERSION") + ")"
		dummyRootLoggerWriter(w, r, msg, http.StatusOK)
	}

	// Simple request logger + response writer.
	dummyRootLoggerWriter = func(w http.ResponseWriter, r *http.Request, msg string, status int) {
		l := common.NewLogger(r, "root")

		// Log this, and write the common response.
		l.Msg(msg).Status(status).Log().Payload(nil).Write(w)
	}
)

// Simple rate limiter (by IP and URL). (Limits Requests per duration, see pkg/config/backend.go for more.)
// https://github.com/go-chi/httprate
var limiter = httprate.Limit(config.LIMITER_REQS_NUM, config.LIMITER_DURATION_SEC*time.Second,
	httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
		// Log the too-many-requests error.
		common.NewLogger(r, "base").Status(http.StatusTooManyRequests).Log()

		// Write simple response.
		http.Error(w, `{"error": "too many requests, slow down"}`, http.StatusTooManyRequests)
	}),
	httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
)

// The very main API router.
func NewAPIRouter() chi.Router {
	r := chi.NewRouter()

	// Use the authentication middleware.
	r.Use(auth.AuthMiddleware)

	// Use the rate limiter, feature-flagged.
	if !config.IsLimiterDisabled {
		r.Use(limiter)
	}

	// Init repositories for services.
	pollRepository := polls.NewPollRepository(db.PollCache)
	postRepository := posts.NewPostRepository(db.FlowCache)
	subscriptionRepository := push.NewSubscriptionRepository(db.SubscriptionCache)
	requestRepository := requests.NewRequestRepository(db.RequestCache)
	tokenRepository := tokens.NewTokenRepository(db.TokenCache)
	userRepository := users.NewUserRepository(db.UserCache)

	// Init services for controllers.
	authService := auth.NewAuthService(tokenRepository, userRepository)
	pollService := polls.NewPollService(pollRepository, postRepository, userRepository)
	postService := posts.NewPostService(postRepository, userRepository)
	statService := stats.NewStatService(pollRepository, postRepository, userRepository)
	userService := users.NewUserService(postRepository, subscriptionRepository, requestRepository, tokenRepository, userRepository)

	// Init controllers for routers.
	authController := auth.NewAuthController(authService)
	pollController := polls.NewPollController(pollService)
	statController := stats.NewStatController(statService)
	userController := users.NewUserController(postService, userService)

	//
	//  API subpkg routers registering
	//

	// Served at /api/v1.
	r.Get("/", rootHandler)

	r.Mount("/auth", auth.NewAuthRouter(authController))

	r.Mount("/dump", db.Router())

	r.Mount("/live", live.NewLiveRouter())
	r.Mount("/polls", polls.NewPollRouter(pollController))

	r.Mount("/posts", posts.Router())
	r.Mount("/push", push.Router())

	r.Mount("/stats", stats.NewStatRouter(statController))
	r.Mount("/users", users.NewUserRouter(userController))

	return r
}
