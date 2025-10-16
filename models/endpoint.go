package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
"strings"
	"time"

"gorm.io/gorm"
)

// IntArray represents a slice of integers that can be stored as JSON in the database
type IntArray []int

func (a IntArray) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *IntArray) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, a)
}

var ErrInvalidEndpoint = errors.New("endpoint requires a non-empty url and positive interval")

type Endpoint struct {
	ID                   string    `gorm:"primaryKey" json:"id"`
	URL                  string    `gorm:"not null" json:"url"`
	Interval             int       `gorm:"not null" json:"interval"` // seconds
	Timeout              int       `gorm:"default:30" json:"timeout"` // seconds, default 30
	ExpectedStatusCodes  IntArray `gorm:"type:json" json:"expected_status_codes"` // empty means 200-299
	MaxResponseTime      int       `gorm:"default:5000" json:"max_response_time"` // milliseconds, default 5000
	CreatedAt            time.Time `json:"created_at"`
}

func (e *Endpoint) BeforeSave(tx *gorm.DB) error {
	e.URL = strings.TrimSpace(e.URL)
	if e.URL == "" || e.Interval <= 0 {
		return ErrInvalidEndpoint
	}

	// Set defaults if not provided
	if e.Timeout <= 0 {
		e.Timeout = 30
	}
	if e.MaxResponseTime <= 0 {
		e.MaxResponseTime = 5000
	}

	return nil
}
