package auth

import (
	chi "github.com/go-chi/chi/v5"
)

func NewAuthRouter(authController *authController) chi.Router {
	r := chi.NewRouter()

	r.Post("/", authController.Auth)
	r.Post("/logout", authController.Logout)

	return r
}
