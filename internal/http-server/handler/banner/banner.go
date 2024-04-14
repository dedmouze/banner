package banner

import (
	storage "banner/internal/database"
	"banner/internal/database/model"
	"banner/internal/http-server/middleware/validator"
	httpBanner "banner/internal/http-server/model"
	"banner/pkg/lib/api/response"
	"banner/pkg/lib/sl"
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/render"
)

type BannerProvider interface {
	BannerByID(ctx context.Context, featureID, tagID int64, limit, offset int64) ([]model.Banner, [][]int64, error)
}

type Response struct {
	response.Response
	Banners []httpBanner.Banner `json:"banners"`
}

func New(log *slog.Logger, bannerProvider BannerProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.Banner.New"

		log := log.With(
			slog.String("op", op),
		)

		log.Info("providing banner")

		req, ok := r.Context().Value(validator.GetBannerKey).(validator.GetBannerRequest)
		if !ok {
			log.Error("failed to convert to request")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ErrServerInternal)
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		banners, bannerTagIDs, err := bannerProvider.BannerByID(r.Context(), req.FeatureID, req.TagID, req.Limit, req.Offset)
		if err != nil {
			if errors.Is(err, storage.ErrBannerNotFound) {
				log.Info("banner not found")
				render.JSON(w, r, response.OK())
			} else {
				log.Error("internal error", sl.Err(err))
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, response.ErrServerInternal)
			}
			return
		}

		if len(banners) != len(bannerTagIDs) {
			log.Error("internal error")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ErrServerInternal)
			return
		}

		var httpBanners []httpBanner.Banner
		for i, banner := range banners {
			httpBanners = append(httpBanners, *httpBanner.BannerDBtoBannerHTTP(banner, req.FeatureID, bannerTagIDs[i]))
		}

		log.Info("banners provided")
		render.JSON(w, r, Response{
			Response: response.OK(),
			Banners:  httpBanners,
		})
	}
}
