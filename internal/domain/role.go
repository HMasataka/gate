package domain

import "time"

type Role struct {
	ID          string
	Name        string
	Description string
	ParentID    *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
