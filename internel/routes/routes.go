package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"url-shortner/internel/http-server/handlers/redirect"
	"url-shortner/internel/http-server/handlers/url/delete"
	"url-shortner/internel/http-server/handlers/url/save"
	"url-shortner/internel/storage/sqlite"
)

func New(log *slog.Logger, storage *sqlite.Storage, user, password string) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortener", map[string]string{
			user: password,
		}))
		r.Post("/", save.New(log, storage))
		r.Delete("/{id}", delete.New(log, storage))
	})

	router.Get("/{alias}", redirect.New(log, storage))

	return router
}
