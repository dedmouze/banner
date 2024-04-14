package create

import (
	"banner/internal/database/model"
	"banner/internal/http-server/middleware/validator"
	"banner/pkg/lib/api/response"
	"banner/pkg/lib/sl"
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/render"
)

type BannerCreator interface {
	CreateBanner(ctx context.Context, banner *model.Banner, feature *model.Feature, tags []model.Tag) (int64, error)
}

type Response struct {
	response.Response
	BannerID int64 `json:"banner_id"`
}

func New(log *slog.Logger, bannerCreator BannerCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.Banner.Create.New"

		log := log.With(
			slog.String("op", op),
		)

		log.Info("creating banner")

		req, ok := r.Context().Value(validator.PostBannerKey).(validator.PostBannerRequest)
		if !ok {
			log.Error("failed convert to request")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ErrServerInternal)
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		content, ok := req.Content["content"].(string)
		if !ok {
			log.Error("failed to convert to request")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ErrServerInternal)
			return
		}

		t := time.Now()
		banner := &model.Banner{
			IsActive:  req.IsActive,
			Content:   content,
			CreatedAt: t,
			UpdatedAt: t,
		}

		feature := &model.Feature{
			ID:        req.FeatureID,
			CreatedAt: t,
			UsedAt:    t,
		}

		tags := make([]model.Tag, len(req.TagIDs))
		for i, v := range req.TagIDs {
			tags[i] = model.Tag{
				ID:        v,
				CreatedAt: t,
				UsedAt:    t,
			}
		}

		log.Debug(
			"decoded parameters",
			slog.Any("banner", banner),
			slog.Any("feature", feature),
			slog.Any("tags", tags),
		)

		id, err := bannerCreator.CreateBanner(r.Context(), banner, feature, tags)
		if err != nil {
			log.Error("internal error", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ErrServerInternal)
			return
		}

		log.Info("banner created")
		render.Status(r, http.StatusCreated)
		render.JSON(w, r, Response{
			Response: response.Created(),
			BannerID: id,
		})
	}
}
