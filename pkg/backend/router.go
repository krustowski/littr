//	@title			littr
//	@version		0.46.51
//	@description		A simple nanoblogging platform.
//	@description
//	@description		HTTP cookies must be used for authentication on most routes. These can be obtained by calling the `/auth` route with the appropriate parameters.
//	@termsOfService		https://www.littr.eu/tos

//	@contact.name		API Support
//	@contact.url		https://www.littr.eu/docs
//	@contact.email		info@littr.eu

//	@license.name		MIT
//	@license.url		https://github.com/krustowski/littr/blob/master/LICENSE

//	@host			www.littr.eu
//	@BasePath		/api/v1
//	@accept			json
//	@produce		json
//	@schemes		https http

//	@supportedSubmitMethods		[]
//	@security			[]
//	@securityDefinitions.basic

//	@externalDocs.description	OpenAPI
//	@externalDocs.url		https://swagger.io/resources/open-api/

//	@externalDocs.description	Documentation
//	@externalDocs.url		https://krusty.space/projects/littr/

//	@tag.name		auth
//	@tag.description	Authentication and HTTP cookies management

//	@tag.name		dump
//	@tag.description	Interventions in running data

//	@tag.name		live
//	@tag.description	Real-time event streaming

//	@tag.name		polls
//	@tag.description	Polls management procedures

//	@tag.name		posts
//	@tag.description	Operations with contributions

//	@tag.name		stats
//	@tag.description	System statistics

//	@tag.name		users
//	@tag.description	User manipulation operations

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
	"go.vxn.dev/littr/pkg/backend/mail"
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
var limiter = httprate.Limit(config.ApiLimiterRequestsCount, time.Second*config.ApiLimiterDuration,
	httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
		l := common.NewLogger(r, "root")

		// Do not log this, just write the response!
		l.Msg("too many requests, slow down and try again later").Status(http.StatusTooManyRequests).Payload(nil).Write(w)
	}),
	httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
)

// The very main API router.
func NewAPIRouter(db db.DatabaseKeeper) chi.Router {
	r := chi.NewRouter()

	// Use the authentication middleware.
	r.Use(auth.AuthMiddleware)

	// Use the rate limiter, feature-flagged.
	if config.IsApiLimiterEnabled {
		r.Use(limiter)
	}

	// Define notFound and methodNotAllowed default handlers.
	r.NotFound(http.HandlerFunc(NotFoundHandler))
	r.MethodNotAllowed(http.HandlerFunc(MethodNotAllowedHandler))

	// Init repositories for services.
	pollRepository := polls.NewPollRepository(db.PollCache)
	postRepository := posts.NewPostRepository(db.FlowCache)
	requestRepository := requests.NewRequestRepository(db.RequestCache)
	tokenRepository := tokens.NewTokenRepository(db.TokenCache)
	userRepository := users.NewUserRepository(db.UserCache)

	// Init services for controllers.
	authService := auth.NewAuthService(tokenRepository, userRepository)
	mailService := mail.NewMailService()
	notifService := push.NewNotificationService(postRepository, userRepository)
	pollService := polls.NewPollService(pollRepository, postRepository, userRepository)
	postService := posts.NewPostService(notifService, postRepository, userRepository)
	statService := stats.NewStatService(pollRepository, postRepository, userRepository)
	userService := users.NewUserService(mailService, pollRepository, postRepository, requestRepository, tokenRepository, userRepository)

	// Init controllers for routers.
	authController := auth.NewAuthController(authService)
	pollController := polls.NewPollController(pollService)
	postController := posts.NewPostController(postService)
	statController := stats.NewStatController(statService)
	userController := users.NewUserController(postService, statService, userService)

	//
	//  API subpkg routers registering
	//

	// Served at /api/v1.
	r.Get("/", rootHandler)

	r.Mount("/auth", auth.NewAuthRouter(authController))
	r.Mount("/dump", db.NewDumpRouter())
	r.Mount("/live", live.NewLiveRouter())
	r.Mount("/polls", polls.NewPollRouter(pollController))
	r.Mount("/posts", posts.NewPostRouter(postController))
	r.Mount("/stats", stats.NewStatRouter(statController))
	r.Mount("/users", users.NewUserRouter(userController))

	return r
}
