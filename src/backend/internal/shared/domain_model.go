package shared

import "time"

type DomainModel struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
}
