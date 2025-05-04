package models

import "time"

type Topic struct {
	ID        int
	Title     string
	Content   string
	UserID    int64
	CreatedAt time.Time
}
