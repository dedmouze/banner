package model

import "time"

type Tag struct {
	ID        int64     `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UsedAt    time.Time `db:"used_at"`
}
