package delete

import (
	storage "banner/internal/database"
	"banner/internal/http-server/middleware/validator"
	"banner/pkg/lib/api/response"
	"banner/pkg/lib/sl"
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/render"
)

type BannerDeleter interface {
	DeleteBanner(ctx context.Context, bannerID int64) error
}

type Response struct {
	response.Response
}

func New(log *slog.Logger, bannerDeleter BannerDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.Banner.Delete.New"

		log := log.With(
			slog.String("op", op),
		)

		log.Info("deleting banner")

		req, ok := r.Context().Value(validator.DeleteBannerWithIDKey).(validator.DeleteBannerWithID)
		if !ok {
			log.Error("failed convert to request")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ErrServerInternal)
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		err := bannerDeleter.DeleteBanner(r.Context(), req.BannerID)
		if err != nil {
			if errors.Is(err, storage.ErrBannerNotFound) {
				log.Info("banner not found")
				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, response.ErrBannerNotFound)
			} else {
				log.Error("internal error", sl.Err(err))
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, response.ErrServerInternal)
			}
			return
		}

		log.Info("banner deleted")
		render.Status(r, http.StatusOK)
		render.JSON(w, r, Response{
			Response: response.OK(),
		})
	}
}
