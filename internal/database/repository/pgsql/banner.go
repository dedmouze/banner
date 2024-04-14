package pgsql

import (
	storage "banner/internal/database"
	"banner/internal/database/model"
	"database/sql"
	"errors"
	"time"

	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type BannerRepository struct {
	db *sqlx.DB
}

func NewBannerRepository(db *sqlx.DB) *BannerRepository {
	return &BannerRepository{db: db}
}

func (b *BannerRepository) Banner(ctx context.Context, featureID, tagID int64) (string, error) {
	const op = "repository.pgsql.Banner"

	stmt, err := b.db.PrepareContext(ctx,
		`
		SELECT b.content FROM banner b
		INNER JOIN banner_feature f ON f.banner_id = b.id AND f.feature_id = $1
		INNER JOIN banner_tag t ON t.banner_id = b.id AND t.tag_id = $2
		LIMIT 1 OFFSET 0
		`,
	)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, featureID, tagID)

	var content string
	if err = row.Scan(&content); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("%s: %w", op, storage.ErrBannerNotFound)
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return content, nil
}

func (b *BannerRepository) BannerByID(ctx context.Context, featureID, tagID int64, limit, offset int64) ([]model.Banner, [][]int64, error) {
	const op = "repository.pgsql.BannerByID"

	stmt1 := "SELECT b.id, b.content, b.is_active, b.created_at, b.updated_at FROM banner b"
	stmt2 := fmt.Sprintf("INNER JOIN banner_feature f ON f.banner_id = b.id AND f.feature_id = %d", featureID)
	stmt3 := fmt.Sprintf("INNER JOIN banner_tag t ON t.banner_id = b.id AND t.tag_id = %d", tagID)
	stmt4 := fmt.Sprintf("LIMIT %d", limit)
	stmt5 := fmt.Sprintf("OFFSET %d", offset)

	var err error
	var stmt *sqlx.Stmt
	if limit != 0 {
		if featureID == 0 && tagID == 0 {
			stmt, err = b.db.PreparexContext(ctx, fmt.Sprintf("%s %s %s", stmt1, stmt4, stmt5))
		} else if featureID != 0 && tagID == 0 {
			stmt, err = b.db.PreparexContext(ctx, fmt.Sprintf("%s %s %s %s", stmt1, stmt2, stmt4, stmt5))
		} else if featureID == 0 && tagID != 0 {
			stmt, err = b.db.PreparexContext(ctx, fmt.Sprintf("%s %s %s %s", stmt1, stmt3, stmt4, stmt5))
		} else if featureID != 0 && tagID != 0 {
			stmt, err = b.db.PreparexContext(ctx, fmt.Sprintf("%s %s %s %s %s", stmt1, stmt2, stmt3, stmt4, stmt5))
		}
	} else {
		if featureID == 0 && tagID == 0 {
			stmt, err = b.db.PreparexContext(ctx, stmt1)
		} else if featureID != 0 && tagID == 0 {
			stmt, err = b.db.PreparexContext(ctx, fmt.Sprintf("%s %s", stmt1, stmt2))
		} else if featureID == 0 && tagID != 0 {
			stmt, err = b.db.PreparexContext(ctx, fmt.Sprintf("%s %s", stmt1, stmt3))
		} else if featureID != 0 && tagID != 0 {
			stmt, err = b.db.PreparexContext(ctx, fmt.Sprintf("%s %s %s", stmt1, stmt2, stmt3))
		}
	}
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := stmt.QueryxContext(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, fmt.Errorf("%s: %w", op, storage.ErrBannerNotFound)
		}
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var bannerTagIDs [][]int64
	var banners []model.Banner
	for rows.Next() {
		var banner model.Banner
		if err := rows.StructScan(&banner); err != nil {
			return nil, nil, fmt.Errorf("%s: %w", op, err)
		}
		banners = append(banners, banner)

		stmt, err := b.db.PrepareContext(ctx, "SELECT t.id FROM tag t INNER JOIN banner_tag bt ON t.id = bt.tag_id AND banner_id = ?")
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", op, err)
		}

		tagRows, err := stmt.QueryContext(ctx, banner.ID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, nil, fmt.Errorf("%s: %w", op, err)
		}
		defer tagRows.Close()

		var tagIDs []int64
		for tagRows.Next() {
			var tagID int64
			if err := tagRows.Scan(&tagID); err != nil {
				return nil, nil, fmt.Errorf("%s: %w", op, err)
			}
			tagIDs = append(tagIDs, tagID)
		}
		if err = rows.Err(); err != nil {
			return nil, nil, fmt.Errorf("%s: %w", op, err)
		}
		bannerTagIDs = append(bannerTagIDs, tagIDs)
	}
	if err = rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	return banners, bannerTagIDs, nil
}

func (b *BannerRepository) CreateBanner(ctx context.Context, banner *model.Banner, feature *model.Feature, tags []model.Tag) (int64, error) {
	const op = "repository.pgsql.CreateBanner"

	var err error
	txx, err := b.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer txx.Rollback()

	var bannerID int64 = 0
	row := txx.QueryRowContext(ctx, "SELECT id FROM banner WHERE content = $1", banner.Content)
	if err := row.Scan(&bannerID); err != nil {
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("%s: %w", op, err)
		}
	}

	if bannerID == 0 {
		err := txx.QueryRowContext(ctx, "INSERT INTO banner (content, is_active, created_at, updated_at) VALUES ($1, $2, $3, $4) RETURNING id",
			banner.Content, banner.IsActive, banner.CreatedAt, banner.UpdatedAt,
		).Scan(&bannerID)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", op, err)
		}
	}

	var featureID int64 = 0
	row = txx.QueryRowContext(ctx, "SELECT id FROM feature WHERE id = $1", feature.ID)
	if err := row.Scan(&featureID); err != nil {
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("%s: %w", op, err)
		}
	}

	if featureID == 0 {
		err := txx.QueryRowContext(ctx, "INSERT INTO feature (id, created_at, used_at) VALUES ($1, $2, $3) RETURNING id",
			feature.ID, feature.CreatedAt, feature.UsedAt,
		).Scan(&featureID)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", op, err)
		}
	}

	var id int64
	err = txx.QueryRowContext(
		ctx, "SELECT banner_id FROM banner_feature WHERE banner_id = $1 AND feature_id = $2",
		bannerID, featureID).Scan(&id)
	if err != nil {
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("%s: %w", op, err)
		}
	}

	if errors.Is(err, sql.ErrNoRows) {
		_, err = txx.ExecContext(ctx, "INSERT INTO banner_feature (banner_id, feature_id) VALUES ($1, $2)",
			bannerID, featureID,
		)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", op, err)
		}
	}

	for _, tag := range tags {
		var tagID int64 = 0
		row = txx.QueryRowContext(ctx, "SELECT id FROM tag WHERE id = $1", tag.ID)
		if err := row.Scan(&tagID); err != nil {
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return 0, fmt.Errorf("%s: %w", op, err)
			}
		}

		if tagID == 0 {
			err := txx.QueryRowContext(ctx, "INSERT INTO tag (id, created_at, used_at) VALUES ($1, $2, $3) RETURNING id",
				tag.ID, tag.CreatedAt, tag.UsedAt,
			).Scan(&tagID)
			if err != nil {
				return 0, fmt.Errorf("%s: %w", op, err)
			}
		}

		err = txx.QueryRowContext(
			ctx, "SELECT banner_id FROM banner_tag WHERE banner_id = $1 AND tag_id = $2",
			bannerID, tagID).Scan(&id)
		if err != nil {
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return 0, fmt.Errorf("%s: %w", op, err)
			}
		}

		if errors.Is(err, sql.ErrNoRows) {
			_, err = txx.ExecContext(ctx, "INSERT INTO banner_tag (banner_id, tag_id) VALUES ($1, $2)",
				bannerID, tagID,
			)
			if err != nil {
				return 0, fmt.Errorf("%s: %w", op, err)
			}
		}
	}

	if err = txx.Commit(); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return bannerID, nil
}

func (b *BannerRepository) UpdateBanner(ctx context.Context, banner *model.Banner, featureID int64, tagsID []int64) error {
	const op = "repository.pgsql.UpdateBanner"

	txx, err := b.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer txx.Rollback()

	stmt, err := txx.PrepareContext(ctx, "UPDATE banner SET content = $1, is_active = $2, updated_at = $3 WHERE id = $4")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.ExecContext(ctx, banner.Content, banner.IsActive, banner.UpdatedAt, banner.ID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrBannerNotFound)
	}

	var newFeatureID int64 = 0
	row := txx.QueryRowContext(ctx, "SELECT id FROM feature WHERE id = $1", featureID)
	if err := row.Scan(&newFeatureID); err != nil {
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	now := time.Now()
	if newFeatureID == 0 {
		err := txx.QueryRowContext(ctx, "INSERT INTO feature (id, created_at, used_at) VALUES ($1, $2, $3) RETURNING id",
			featureID, now, now,
		).Scan(&featureID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	stmt, err = txx.PrepareContext(ctx, "UPDATE banner_feature SET feature_id = $1 WHERE banner_id = $2")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	_, err = stmt.ExecContext(ctx, featureID, banner.ID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	stmt, err = txx.PrepareContext(ctx, "DELETE FROM banner_tag WHERE banner_id = $1")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	_, err = stmt.ExecContext(ctx, banner.ID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	for _, tagID := range tagsID {
		var newTagID int64 = 0
		row := txx.QueryRowContext(ctx, "SELECT id FROM tag WHERE id = $1", tagID)
		if err := row.Scan(&newTagID); err != nil {
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("%s: %w", op, err)
			}
		}

		if newTagID == 0 {
			err := txx.QueryRowContext(ctx, "INSERT INTO tag (id, created_at, used_at) VALUES ($1, $2, $3) RETURNING id",
				tagID, now, now,
			).Scan(&featureID)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
		}

		stmt, err = txx.PrepareContext(ctx, "INSERT INTO banner_tag (banner_id, tag_id) VALUES($1, $2)")
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		_, err := stmt.ExecContext(ctx, banner.ID, tagID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	if err = txx.Commit(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (b *BannerRepository) DeleteBanner(ctx context.Context, bannerID int64) error {
	const op = "repository.pgsql.DeleteBanner"

	tx, err := b.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	err = handleDelete(ctx, tx, bannerID, "DELETE FROM banner_tag WHERE banner_id = $1", storage.ErrBannerTagRelationNotFound)
	if err != nil && !errors.Is(err, storage.ErrBannerTagRelationNotFound) {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = handleDelete(ctx, tx, bannerID, "DELETE FROM banner_feature WHERE banner_id = $1", storage.ErrBannerFeatureRelationNotFound)
	if err != nil && !errors.Is(err, storage.ErrBannerFeatureRelationNotFound) {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = handleDelete(ctx, tx, bannerID, "DELETE FROM banner WHERE id = $1", storage.ErrBannerNotFound)
	if err != nil {
		return fmt.Errorf("%s: %w", op, storage.ErrBannerNotFound)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func handleDelete(ctx context.Context, tx *sql.Tx, id int64, queryDelete string, ErrNotFound error) error {
	const op = "repository.pgsql.handleDelete"

	stmt, err := tx.PrepareContext(ctx, queryDelete)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	res, err := stmt.ExecContext(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	affectedRows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if affectedRows == 0 {
		return fmt.Errorf("%s: %w", op, ErrNotFound)
	}

	return nil
}
