package repository

import (
	"banner/internal/database/model"
	"context"
)

type BannerRepository interface {
	CreateBanner(context.Context, *model.Banner, *model.Feature, []model.Tag) (int64, error)
	UpdateBanner(context.Context, *model.Banner, *model.Feature, []model.Tag) error
	DeleteBanner(ctx context.Context, bannerID int64) error
	Banner(ctx context.Context, featureID, tagID int64) (*model.Banner, error)
	BannerByID(ctx context.Context, featureID, tagID int64, limit, offset int64) ([]model.Banner, error)
}
