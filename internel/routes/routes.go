package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"log/slog"
	"url-shortner/internel/http-server/handlers/auth/login"
	"url-shortner/internel/http-server/handlers/redirect"
	"url-shortner/internel/http-server/handlers/url/delete"
	"url-shortner/internel/http-server/handlers/url/save"
	"url-shortner/internel/lib/auth/jwt"
	"url-shortner/internel/storage/sqlite"
)

func New(log *slog.Logger, storage *sqlite.Storage) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/auth", func(r chi.Router) {
		r.Post("/login", login.New(log, storage))
	})

	router.Route("/url", func(r chi.Router) {
		r.Use(jwtauth.Verifier(jwt.TokenAuth))
		r.Use(jwtauth.Authenticator(jwt.TokenAuth))

		r.Post("/", save.New(log, storage))
		r.Delete("/{id}", delete.New(log, storage))
	})

	router.Get("/{alias}", redirect.New(log, storage))

	return router
}
