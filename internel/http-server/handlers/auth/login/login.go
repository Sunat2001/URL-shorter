package login

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	"url-shortner/internel/domain/entities/user"
	"url-shortner/internel/lib/api/response"
	"url-shortner/internel/lib/auth/authRequest"
	"url-shortner/internel/lib/auth/authResponse"
	"url-shortner/internel/lib/auth/hash"
	"url-shortner/internel/lib/auth/jwt"
	"url-shortner/internel/lib/logger/sl"
	"url-shortner/internel/storage"
)

type UserRepository interface {
	GetUser(userName string) (user.User, error)
}

func New(log *slog.Logger, userRepository UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.auth.login.New"

		log.With(
			slog.String("op", op),
			slog.String("requestId", middleware.GetReqID(r.Context())),
		)

		var req authRequest.Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("Failed to render.JSON", sl.Err(err))
			render.JSON(w, r, response.Error("failed parse JSON body"))
			return
		}

		log.Info("request Body decoded", slog.Any("request", req))
		if err := validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)
			log.Error("Invalid request", sl.Err(err))
			render.JSON(w, r, response.ValidationError(validateErr))
			return
		}

		user, err := userRepository.GetUser(req.Username)
		if errors.Is(err, storage.UserNotFound) {
			log.Info("User doesn't exist", slog.String("username", req.Username))
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, response.Error("Invalid username or password"))
			return
		}
		if err != nil {
			log.Error("Failed to add url", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("Internal Server Error"))
			return
		}

		if !hash.CheckPasswordHash(req.Password, user.Password) {
			log.Info("Invalid password", slog.String("password", req.Password))
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, response.Error("Invalid username or password"))
			return
		}

		token, err := jwt.GenerateToken(user.ID)
		if err != nil {
			log.Error("Failed to generate jwt", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("Internal Server Error"))
			return
		}
		responseOK(w, r, user, token)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, User user.User, token string) {
	render.JSON(w, r, authResponse.Response{
		Response: response.OK(),
		User: user.User{
			ID:       User.ID,
			Username: User.Username,
		},
		AuthTokenInfo: authResponse.AuthTokenInfo{
			Token: token,
			Type:  "bearer",
		},
	})
}
