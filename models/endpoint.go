package models

import "time"

type Endpoint struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	URL       string    `json:"url"`
	Interval  int       `json:"interval"` // seconds
	CreatedAt time.Time `json:"created_at"`
}
