package storage

import "errors"

var (
	ErrBannerNotFound                = errors.New("banner not found")
	ErrBannerAlreadyExists           = errors.New("banner already exists")
	ErrFeatureAlredyExists           = errors.New("feature already exists")
	ErrTagAlreadyExists              = errors.New("tag already exists")
	ErrBannerTagRelationNotFound     = errors.New("banner-tag relation not found")
	ErrBannerFeatureRelationNotFound = errors.New("banner-feature relation not found")
)
