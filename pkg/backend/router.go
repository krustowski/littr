// @title		littr
// @version	 	0.39.12
// @description		nanoblogging platform as PWA built on go-app framework
// @termsOfService	https://littr.eu/tos

// @contact.name	API Support
// @contact.url		https://littr.eu/docs/
// @contact.email	info@littr.eu

// @license.name	MIT
// @license.url		https://github.com/krustowski/littr/blob/master/LICENSE

// @host		littr.eu
// @BasePath		/api/v1

// @supportedSubmitMethods	[]
// @security            	[]
// @securityDefinitions.basic	BasicAuth

// @externalDocs.description	OpenAPI
// @externalDocs.url		https://swagger.io/resources/open-api/
package backend

import (
	"net/http"
	"os"

	chi "github.com/go-chi/chi/v5"

	"go.vxn.dev/littr/pkg/backend/auth"
	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/backend/db"
	"go.vxn.dev/littr/pkg/backend/polls"
	"go.vxn.dev/littr/pkg/backend/posts"
	"go.vxn.dev/littr/pkg/backend/push"
	"go.vxn.dev/littr/pkg/backend/stats"
	"go.vxn.dev/littr/pkg/backend/users"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	resp := common.Response{
		Message: "littr API service (v" + os.Getenv("APP_VERSION") + ")",
		Code:    http.StatusOK,
	}
	resp.Write(w)
}

// the very main API router
func APIRouter() chi.Router {
	r := chi.NewRouter()

	r.Use(auth.AuthMiddleware)
	//r.Use(system.LoggerMiddleware)

	r.Get("/", rootHandler)

	r.Mount("/auth", auth.Router())
	r.Mount("/dump", db.Router())
	r.Mount("/polls", polls.Router())
	r.Mount("/posts", posts.Router())
	r.Mount("/push", push.Router())
	r.Mount("/stats", stats.Router())
	r.Mount("/users", users.Router())

	return r
}
