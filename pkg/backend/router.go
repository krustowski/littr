// @title		litter-go
// @version	 	0.33.0
// @description	nanoblogging platform as PWA built on go-app framework (PoC)
// @termsOfService	https://littr.n0p.cz/tos

// @contact.name	API Support
// @contact.url		https://littr.n0p.cz/docs/
// @contact.email	littr@n0p.cz

// @license.name	MIT
// @license.url		https://github.com/krustowski/litter-go/blob/master/LICENSE

// @host		littr.n0p.cz
// @BasePath		/api/v1

// @securityDefinitions.basic	BasicAuth

// @externalDocs.description	OpenAPI
// @externalDocs.url		https://swagger.io/resources/open-api/
package backend

import (
	"context"
	"net/http"
	"os"
	"time"

	chi "github.com/go-chi/chi/v5"

	"go.savla.dev/littr/pkg/backend/auth"
	"go.savla.dev/littr/pkg/backend/data"
	"go.savla.dev/littr/pkg/backend/polls"
	"go.savla.dev/littr/pkg/backend/posts"
	"go.savla.dev/littr/pkg/backend/stats"
	"go.savla.dev/littr/pkg/backend/users"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	resp := Response{
		Message: "litter-go API service (v"+os.Getenv("APP_VERSION")+")",
		Code:    http.StatusOK,
	}
	resp.Write(w)
}

// the very main API router
func APIRouter() chi.Router {
	r := chi.NewRouter()

	r.Use(auth.AuthMiddleware)
	r.Use(system.LoggerMiddleware)

	r.Get("/", rootHandler)

	r.Mount("/auth", auth.Router())
	r.Mount("/dump", db.Router())
	r.Mount("/polls", polls.Router())
	r.Mount("/posts", posts.Router())
	r.Mount("/stats", stats.Router())
	r.Mount("/users", users.Router())

	return r
}
