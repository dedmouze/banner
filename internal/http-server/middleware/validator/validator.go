package validator

import (
	"banner/pkg/lib/api/response"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/render"
)

const (
	userBanner = "/user_banner"
	banner     = "/banner"
)

func New(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		const op = "http-server.middleware.validator"

		log := log.With(
			slog.String("op", op),
		)

		log.Info("validator middleware enabled")

		fn := func(w http.ResponseWriter, r *http.Request) {

			notImplemented := false

			var (
				ctx context.Context
				err error
				ok  bool
			)
			path := r.URL.Path
			method := r.Method
			if path == userBanner {
				ok, ctx, err = validateUserBanner(r)
				ok = validate(ok, err, &w, r, log)
				if !ok {
					return
				}
			} else if path == banner {
				if method == http.MethodGet || method == http.MethodPost {
					ok, ctx, err = validateBanner(r)
					ok = validate(ok, err, &w, r, log)
					if !ok {
						return
					}
				} else {
					notImplemented = true
				}
			} else {
				if method == http.MethodPatch || method == http.MethodDelete {
					ok, ctx, err = validateBannerWithID(r)
					ok = validate(ok, err, &w, r, log)
					if !ok {
						return
					}
				} else {
					notImplemented = true
				}
			}

			if notImplemented {
				log.Info("not implemented", slog.String("path", path), slog.String("method", method))
				render.Status(r, http.StatusNotImplemented)
				render.JSON(w, r, response.Error(response.ErrNotImplemented.Error()))
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}

func validate(ok bool, err error, w *http.ResponseWriter, r *http.Request, log *slog.Logger) bool {
	if err != nil {
		log.Error("internal error")
		render.Status(r, http.StatusInternalServerError)
		render.JSON(*w, r, response.ErrServerInternal)
		return false
	}
	if !ok {
		log.Info("bad request")
		render.Status(r, http.StatusBadRequest)
		render.JSON(*w, r, response.ErrBadRequest)
		return false
	}
	return true
}

var bannerQueryParams = []string{
	"feature_id",
	"tag_id",
	"limit",
	"offset",
}

type PostBannerRequest struct {
	FeatureID int64                  `json:"feature_id"`
	TagIDs    []int64                `json:"tag_ids"`
	Content   map[string]interface{} `json:"content"`
	IsActive  bool                   `json:"is_active"`
}

type GetBannerRequest struct {
	FeatureID int64
	TagID     int64
	Limit     int64
	Offset    int64
}

type Key string

const (
	GetBannerKey  = Key("get banner key")
	PostBannerKey = Key("post banner key")
)

func validateBanner(r *http.Request) (bool, context.Context, error) {
	var ctx context.Context
	if r.Method == http.MethodGet {
		query := r.URL.Query()
		var req GetBannerRequest
		for _, param := range bannerQueryParams {
			if query.Has(param) {
				num, err := strconv.Atoi(query.Get(param))
				if err != nil {
					return false, ctx, response.ErrServerInternal
				}
				if num < 0 {
					return false, ctx, response.ErrBadRequest
				}
				if param == "feature_id" {
					req.FeatureID = int64(num)
				} else if param == "tag_id" {
					req.TagID = int64(num)
				} else if param == "limit" {
					req.Limit = int64(num)
				} else if param == "offset" {
					req.Offset = int64(num)
				}
			}
		}

		ctx = context.WithValue(r.Context(), GetBannerKey, req)

	} else if r.Method == http.MethodPost {
		var req PostBannerRequest
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			return false, ctx, err
		}

		content, err := json.Marshal(req.Content)
		if err != nil {
			return false, ctx, err
		}

		req.Content = map[string]interface{}{
			"content": string(content),
		}

		ctx = context.WithValue(r.Context(), PostBannerKey, req)
	} else {
		return false, ctx, response.ErrNotImplemented
	}

	return true, ctx, nil
}

type GetUserBannerRequest struct {
	FeatureID       int64
	TagID           int64
	UseLastRevision bool
}

const GetUserBannerKey = Key("get user banner key")

var userBannerQueryParams = []string{
	"tag_id",
	"feature_id",
	"use_last_revisison",
}

func validateUserBanner(r *http.Request) (bool, context.Context, error) {
	var ctx context.Context
	var req GetUserBannerRequest
	query := r.URL.Query()
	for _, param := range userBannerQueryParams {
		if query.Has(param) {
			if param == "feature_id" || param == "tag_id" {
				num, err := strconv.Atoi(query.Get(param))
				if err != nil {
					return false, ctx, err
				}
				if param == "feature_id" {
					req.FeatureID = int64(num)
				} else {
					req.TagID = int64(num)
				}
			} else if param == "use_last_revision" {
				if query.Get(param) == "true" {
					req.UseLastRevision = true
				}
			}
		}
	}

	ctx = context.WithValue(r.Context(), GetUserBannerKey, req)
	return true, ctx, nil
}

type PatchBannerWithID struct {
	BannerID  int64
	FeatureID int64                  `json:"feature_id"`
	TagIDs    []int64                `json:"tag_ids"`
	Content   map[string]interface{} `json:"content"`
	IsActive  bool                   `json:"is_active"`
}

type DeleteBannerWithID struct {
	BannerID int64 `json:"banner_id"`
}

const (
	DeleteBannerWithIDKey = Key("delete banner with id")
	PatchBannerWithIDKey  = Key("patch banner with id")
)

func validateBannerWithID(r *http.Request) (bool, context.Context, error) {
	var ctx context.Context

	path := r.URL.Path
	param := path[strings.LastIndex(path, "/")+1:]
	id, err := strconv.Atoi(param)
	if err != nil {
		return false, ctx, response.ErrServerInternal
	}

	if r.Method == http.MethodDelete {
		var req DeleteBannerWithID
		req.BannerID = int64(id)
		ctx = context.WithValue(r.Context(), DeleteBannerWithIDKey, req)
	} else if r.Method == http.MethodPatch {
		var req PatchBannerWithID
		req.BannerID = int64(id)
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			return false, ctx, err
		}

		content, err := json.Marshal(req.Content)
		if err != nil {
			return false, ctx, err
		}

		req.Content = map[string]interface{}{
			"content": string(content),
		}

		ctx = context.WithValue(r.Context(), PatchBannerWithIDKey, req)
	}

	return true, ctx, nil
}
