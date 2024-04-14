package userBanner

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

type BannerContentProvider interface {
	Banner(ctx context.Context, featureID, tagID int64) (string, error)
}

type Response struct {
	response.Response
	Content string `json:"content"`
}

func New(log *slog.Logger, bannerContentProvider BannerContentProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.Banner.userBanner.New"

		log := log.With(
			slog.String("op", op),
		)

		log.Info("providing banner")

		req, ok := r.Context().Value(validator.GetUserBannerKey).(validator.GetUserBannerRequest)
		if !ok {
			log.Error("failed to convert to request")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ErrServerInternal)
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		//TODO: В зависимости от UseLastRevision обращаться или к PSQL или к Redis

		content, err := bannerContentProvider.Banner(r.Context(), req.FeatureID, req.TagID)
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

		log.Info("banner content provided")
		render.JSON(w, r, Response{
			Response: response.OK(),
			Content:  content,
		})
	}
}
