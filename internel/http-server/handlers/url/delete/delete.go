package delete

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"url-shortner/internel/lib/api/response"
	"url-shortner/internel/lib/logger/sl"
	"url-shortner/internel/storage"
)

type Response struct {
	response.Response
}

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLDeleter
type URLDeleter interface {
	DeleteURL(alias string) error
}

func New(log *slog.Logger, deleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.new"

		log.With(
			slog.String("op", op),
			slog.String("requestId", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias is empty")

			render.JSON(w, r, response.Error("invalid request"))

			return
		}

		err := deleter.DeleteURL(alias)
		if errors.Is(err, storage.ErrIdNotFound) {
			log.Error("ID doesn't exist", sl.Err(err))

			render.JSON(w, r, response.Error("ID doesn't exist"))

			return
		} else if err != nil {
			log.Error("falied to delete URL", sl.Err(err))
			render.JSON(w, r, response.Error("falied to delete URL"))
			return
		}

		responseOK(w, r)
	}

}

func responseOK(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, Response{
		Response: response.OK(),
	})
}
