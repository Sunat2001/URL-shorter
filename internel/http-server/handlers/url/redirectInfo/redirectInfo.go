package redirectInfo

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	"url-shortner/internel/lib/api/response"
	"url-shortner/internel/lib/logger/sl"
	"url-shortner/internel/storage/sqlite"
)

type Request struct {
	Start  int `json:"start" validate:"required,number,min=1"`
	Length int `json:"length" validate:"required,number,min=1"`
}

type Response struct {
	response.Response
	URLs []sqlite.RedirectInfo `json:"urlInfo"`
}

type InfoRepository interface {
	GetAllRedirectInfo(start, length int64) ([]sqlite.RedirectInfo, error)
}

func New(log *slog.Logger, repository InfoRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.RedirectInfo.New"

		log.With(
			slog.String("op", op),
			slog.String("requestId", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("falied to render.JSON", sl.Err(err))
			render.JSON(w, r, response.Error("failed parse JSON body"))
			return
		}

		log.Info("request Body decoded", slog.Any("request", req))
		if err := validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)
			log.Error("invalid request", sl.Err(err))
			render.JSON(w, r, response.ValidationError(validateErr))
			return
		}

		infos, err := repository.GetAllRedirectInfo(int64(req.Start), int64(req.Length))
		if err != nil {
			log.Error("Failed to get url Infos", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("Internal Server Error"))
			return
		}

		responseOK(w, r, infos)
	}
}
func responseOK(w http.ResponseWriter, r *http.Request, Urls []sqlite.RedirectInfo) {
	render.JSON(w, r, Response{
		Response: response.OK(),
		URLs:     Urls,
	})
}
