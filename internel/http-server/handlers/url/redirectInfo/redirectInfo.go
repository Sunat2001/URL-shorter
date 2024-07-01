package redirectInfo

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	"strconv"
	"url-shortner/internel/domain/entities/redirectInfo"
	"url-shortner/internel/lib/api/response"
	"url-shortner/internel/lib/logger/sl"
)

type Request struct {
	Start  int `json:"start" validate:"required,number,min=1"`
	Length int `json:"length" validate:"required,number,min=1"`
}

type Response struct {
	response.Response
	URLs []redirectInfo.RedirectInfo `json:"urlInfo"`
}

type IPApiResponse struct {
	Country     string `json:"country"`
	City        string `json:"city"`
	CountryCode string `json:"countryCode"`
}

type InfoRepository interface {
	GetAllRedirectInfo(start, length int64) ([]redirectInfo.RedirectInfo, error)
}

func New(log *slog.Logger, repository InfoRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.RedirectInfo.New"

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

		infos, err := repository.GetAllRedirectInfo(int64(req.Start), int64(req.Length))
		if err != nil {
			log.Error("Failed to get url Infos", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("Internal Server Error"))
			return
		}

		for i, info := range infos {
			country, err := getCountryByIP(info.Ip)
			if err != nil {
				log.Error("Failed to get ip country info", sl.Err(err))
				continue
			}

			infos[i].Country = country.Country
			infos[i].City = country.City
			infos[i].CountryCode = country.CountryCode
		}

		responseOK(w, r, infos)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, Urls []redirectInfo.RedirectInfo) {
	render.JSON(w, r, Response{
		Response: response.OK(),
		URLs:     Urls,
	})
}

func getCountryByIP(ip string) (IPApiResponse, error) {
	url := fmt.Sprintf("http://ip-api.com/json/%s", ip)
	resp, err := http.Get(url)
	if err != nil {
		return IPApiResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return IPApiResponse{}, fmt.Errorf("failed to get response: %s", resp.Status)
	}

	var result IPApiResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return IPApiResponse{}, err
	}

	return result, nil
}
