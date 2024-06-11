package delete

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"strconv"
	"url-shortner/internel/lib/api/response"
	"url-shortner/internel/lib/logger/sl"
	"url-shortner/internel/storage"
)

type Response struct {
	response.Response
}

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLDeleter
type URLDeleter interface {
	DeleteURL(id int64) error
}

func New(log *slog.Logger, deleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.new"

		log.With(
			slog.String("op", op),
			slog.String("requestId", middleware.GetReqID(r.Context())),
		)

		IDString := chi.URLParam(r, "id")
		if IDString == "" {
			log.Info("ID is empty")

			render.JSON(w, r, response.Error("invalid request"))

			return
		}

		ID, err := strconv.ParseInt(IDString, 10, 64)
		if err != nil {
			log.Info("ID is invalid")
			render.JSON(w, r, response.Error("invalid request"))
			return
		}

		err = deleter.DeleteURL(ID)
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
