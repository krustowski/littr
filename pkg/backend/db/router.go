// The very core database/cache data operations package.
package db

import (
	chi "github.com/go-chi/chi/v5"
)

func NewDumpRouter(controller *dumpController) chi.Router {
	r := chi.NewRouter()

	r.Get("/", controller.DumpAll)

	return r
}
