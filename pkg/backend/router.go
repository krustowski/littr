// @title		littr
// @version	 	0.44.16
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
	//"log/slog"
	"net/http"
	"os"
	"time"

	//"github.com/go-chi/chi/v5/middleware"
	//"github.com/go-chi/httplog/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"

	"go.vxn.dev/littr/pkg/backend/auth"
	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/backend/db"
	//"go.vxn.dev/littr/pkg/backend/docs"
	"go.vxn.dev/littr/pkg/backend/live"
	"go.vxn.dev/littr/pkg/backend/polls"
	"go.vxn.dev/littr/pkg/backend/posts"
	"go.vxn.dev/littr/pkg/backend/push"
	"go.vxn.dev/littr/pkg/backend/stats"
	"go.vxn.dev/littr/pkg/backend/users"
	"go.vxn.dev/littr/pkg/config"
)

// Custom Logger structure.
/*var Logger = httplog.NewLogger("littr-logger", httplog.Options{
	LogLevel: slog.LevelDebug,
	JSON:     true,
	Concise:  true,
	//RequestHeaders: true,
	//ResponseHeaders: true,
	MessageFieldName: "message",
	LevelFieldName:   "severity",
	TimeFieldFormat:  time.RFC3339,
	Tags: map[string]string{
		"version": os.Getenv("APP_VERSION"),
		"env":     config.AppEnvironment,
	},
	QuietDownRoutes: []string{
		"/",
		"/ping",
	},
	//QuietDownPeriods: 10 * time.Second,
	SourceFieldName: "source",
})*/

// The JSON API service root path handler.
func rootHandler(w http.ResponseWriter, r *http.Request) {
	// Log this route as well.
	common.NewLogger(r, "root").Status(http.StatusOK).Log()

	// Write common response.
	common.WriteResponse(w, common.APIResponse{
		Message:   "littr JSON API service (v" + os.Getenv("APP_VERSION") + ")",
		Timestamp: time.Now().UnixNano(),
	}, http.StatusOK)
}

// Simple rate limiter (by IP and URL). Requests per duration.
// https://github.com/go-chi/httprate
var limiter = httprate.Limit(33, 10*time.Second,
	httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
		// Log the too-many-requests error.
		common.NewLogger(r, "root").Status(http.StatusTooManyRequests).Log()

		// Write simple response.
		http.Error(w, `{"error": "too many requests, slow down"}`, http.StatusTooManyRequests)
	}),
	httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
)

// The very main API router.
func APIRouter() chi.Router {
	r := chi.NewRouter()

	// Authentication middleware.
	r.Use(auth.AuthMiddleware)

	// Rate limiter, feature-flagged.
	if !config.IsLimiterDisabled {
		r.Use(limiter)
	}

	r.Get("/", rootHandler)

	r.Mount("/auth", auth.Router())
	//r.Mount("/docs", docs.Router())
	r.Mount("/dump", db.Router())
	r.Mount("/live", live.Router())
	r.Mount("/polls", polls.Router())
	r.Mount("/posts", posts.Router())
	r.Mount("/push", push.Router())
	r.Mount("/stats", stats.Router())
	r.Mount("/users", users.Router())

	return r
}
