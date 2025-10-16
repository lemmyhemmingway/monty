package models

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

var ErrInvalidEndpoint = errors.New("endpoint requires a non-empty url and positive interval")

type Endpoint struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	URL       string    `gorm:"not null" json:"url"`
	Interval  int       `gorm:"not null" json:"interval"` // seconds
	CreatedAt time.Time `json:"created_at"`
}

func (e *Endpoint) BeforeSave(tx *gorm.DB) error {
	e.URL = strings.TrimSpace(e.URL)
	if e.URL == "" || e.Interval <= 0 {
		return ErrInvalidEndpoint
	}
	return nil
}
