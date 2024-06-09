package register

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	"url-shortner/internel/lib/api/response"
	"url-shortner/internel/lib/auth/authRequest"
	"url-shortner/internel/lib/logger/sl"
	"url-shortner/internel/storage/sqlite"
)

type UserRepository interface {
	UserSave(userReq authRequest.Request) (error, sqlite.User)
}

// New Deprecated
func New(log *slog.Logger, userRepo UserRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.auth.register.New"

		log.With(
			slog.String("op", op),
			slog.String("requestId", middleware.GetReqID(r.Context())),
		)

		var req authRequest.Request

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

	})
}
