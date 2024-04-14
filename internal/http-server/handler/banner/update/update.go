package update

import (
	storage "banner/internal/database"
	"banner/internal/database/model"
	"banner/internal/http-server/middleware/validator"
	"banner/pkg/lib/api/response"
	"banner/pkg/lib/sl"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/render"
)

type BannerUpdater interface {
	UpdateBanner(ctx context.Context, banner *model.Banner, featureID int64, tagsID []int64) error
}

type Response struct {
	response.Response
}

func New(log *slog.Logger, bannerUpdater BannerUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.Banner.Create.New"

		log := log.With(
			slog.String("op", op),
		)

		log.Info("deleting banner")

		req, ok := r.Context().Value(validator.PatchBannerWithIDKey).(validator.PatchBannerWithID)
		if !ok {
			log.Error("failed convert to request")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ErrServerInternal)
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		content, ok := req.Content["content"].(string)
		if !ok {
			log.Error("failed convert to request")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ErrServerInternal)
			return
		}

		now := time.Now()
		banner := &model.Banner{
			ID:        req.BannerID,
			Content:   content,
			UpdatedAt: now,
			IsActive:  req.IsActive,
		}

		err := bannerUpdater.UpdateBanner(r.Context(), banner, req.FeatureID, req.TagIDs)
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

		log.Info("banner updated")
		render.Status(r, http.StatusOK)
		render.JSON(w, r, Response{
			Response: response.OK(),
		})
	}
}
