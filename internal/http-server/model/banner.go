package model

import (
	"banner/internal/database/model"
	"time"
)

type Banner struct {
	ID        int64     `json:"banner_id"`
	TagIDs    []int64   `json:"tag_ids"`
	FeatureID int64     `json:"feature_id"`
	Content   string    `json:"content"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func BannerDBtoBannerHTTP(banner model.Banner, featureID int64, tagIDs []int64) *Banner {
	return &Banner{
		ID:        banner.ID,
		Content:   banner.Content,
		IsActive:  banner.IsActive,
		CreatedAt: banner.CreatedAt,
		UpdatedAt: banner.UpdatedAt,
		FeatureID: featureID,
		TagIDs:    tagIDs,
	}
}
