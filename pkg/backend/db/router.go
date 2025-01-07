// The very core database/cache data operations package.
package db

import (
	chi "github.com/go-chi/chi/v5"
)

func NewDumpRouter() chi.Router {
	r := chi.NewRouter()

	r.Get("/", dumpHandler)

	return r
}
