package model

import (
	"time"
)

type Banner struct {
	ID        int64     `db:"id"`
	Content   string    `db:"content"`
	IsActive  bool      `db:"is_active"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
