package all

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	"strconv"
	"url-shortner/internel/domain/entities/urlInfo"
	"url-shortner/internel/lib/api/response"
	"url-shortner/internel/lib/logger/sl"
)

type Request struct {
	Start  int `json:"start" validate:"required,number,min=1"`
	Length int `json:"length" validate:"required,number,min=1"`
}

type Response struct {
	response.Response
	URLs []urlInfo.UrlInfo `json:"urls"`
}

type UrlRepository interface {
	GetAllUrl(start, length int64) ([]urlInfo.UrlInfo, error)
}

func New(log *slog.Logger, repository UrlRepository) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.all.New"

		log.With(
			slog.String("op", op),
			slog.String("requestId", middleware.GetReqID(r.Context())),
		)

		startStr := r.URL.Query().Get("start")
		lengthStr := r.URL.Query().Get("length")

		start, err := strconv.Atoi(startStr)
		if err != nil {
			render.JSON(w, r, response.Error("Invalid 'start' parameter"))
			return
		}

		length, err := strconv.Atoi(lengthStr)
		if err != nil {
			render.JSON(w, r, response.Error("Invalid 'length' parameter"))
			return
		}

		var req Request
		req.Start = start
		req.Length = length

		log.Info("request Body decoded", slog.Any("request", req))
		if err := validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)
			log.Error("invalid request", sl.Err(err))
			render.JSON(w, r, response.ValidationError(validateErr))
			return
		}

		urls, err := repository.GetAllUrl(int64(req.Start), int64(req.Length))
		if err != nil {
			log.Error("Failed to get  urls", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("Internal Server Error"))
			return
		}

		responseOK(w, r, urls)
	})
}
func responseOK(w http.ResponseWriter, r *http.Request, Urls []urlInfo.UrlInfo) {
	render.JSON(w, r, Response{
		Response: response.OK(),
		URLs:     Urls,
	})
}
