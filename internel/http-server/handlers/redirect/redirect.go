package redirect

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/mssola/useragent"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"url-shortner/internel/domain/entities/redirectInfo"
	"url-shortner/internel/lib/api/response"
	"url-shortner/internel/lib/logger/sl"
	"url-shortner/internel/storage"
)

// URLGetter is an interface for getting url by alias.
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLGetter
type URLGetter interface {
	GetURL(alias string) (string, error)
	SaveRedirectInfo(redirectInfo *redirectInfo.RedirectInfo) error
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.redirect.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias is empty")

			render.JSON(w, r, response.Error("invalid request"))

			return
		}

		userAgentString := r.Header.Get("User-Agent")
		ua := useragent.New(userAgentString)
		name, version := ua.Browser()
		browser := name + " " + version
		redirectInfoEntity := &redirectInfo.RedirectInfo{
			Ip:       getIP(r),
			Os:       ua.OS(),
			Platform: ua.Platform(),
			Browser:  browser,
		}
		err := urlGetter.SaveRedirectInfo(redirectInfoEntity)
		if err != nil {
			log.Error("Failed to save redirect info", sl.Err(err))

			render.JSON(w, r, response.Error("internal error"))

			return
		}

		resURL, err := urlGetter.GetURL(alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not found", "alias", alias)

			render.JSON(w, r, response.Error("not found"))

			return
		}
		if err != nil {
			log.Error("failed to get url", sl.Err(err))

			render.JSON(w, r, response.Error("internal error"))

			return
		}

		log.Info("got url", slog.String("url", resURL))

		// redirect to found url
		http.Redirect(w, r, resURL, http.StatusFound)
	}
}

func getIP(r *http.Request) string {
	// Check for X-Forwarded-For header (common when behind proxies)
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(ip, ",")
		return strings.TrimSpace(ips[0])
	}

	// Fallback to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}
