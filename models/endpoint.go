package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
"strings"
	"time"

"gorm.io/gorm"
)

// StringArray represents a slice of strings that can be stored as JSON in the database
type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *StringArray) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, a)
}

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
	ID                   string      `gorm:"primaryKey" json:"id"`
	URL                  string      `gorm:"not null" json:"url"`
	CheckType            string      `gorm:"default:http" json:"check_type"` // "http", "ssl", "domain"
	Interval             int         `gorm:"not null" json:"interval"` // seconds
	Timeout              int         `gorm:"default:30" json:"timeout"` // seconds, default 30
	ExpectedStatusCodes  IntArray   `gorm:"type:json" json:"expected_status_codes"` // empty means 200-299
	MaxResponseTime      int         `gorm:"default:5000" json:"max_response_time"` // milliseconds, default 5000
	// SSL-specific fields
	MinDaysValid         int         `gorm:"default:30" json:"min_days_valid"` // days, default 30
	CheckChain           bool        `gorm:"default:true" json:"check_chain"` // default true
	CheckDomainMatch     bool        `gorm:"default:true" json:"check_domain_match"` // default true
	AcceptableTLSVersions StringArray `gorm:"type:json" json:"acceptable_tls_versions"` // e.g., ["TLS 1.2", "TLS 1.3"]
	CreatedAt            time.Time   `json:"created_at"`
}

func (e *Endpoint) BeforeSave(tx *gorm.DB) error {
	e.URL = strings.TrimSpace(e.URL)
	if e.URL == "" || e.Interval <= 0 {
		return ErrInvalidEndpoint
	}

	// Set check type default if not provided
	if e.CheckType == "" {
		e.CheckType = "http"
	}

	// Set defaults if not provided
	if e.Timeout <= 0 {
		e.Timeout = 30
	}
	if e.MaxResponseTime <= 0 {
		e.MaxResponseTime = 5000
	}

	// SSL-specific defaults
	if e.CheckType == "ssl" {
		if e.Interval == 60 { // if default interval, set to 24h for SSL
			e.Interval = 86400
		}
		if e.MinDaysValid <= 0 {
			e.MinDaysValid = 30
		}
		if len(e.AcceptableTLSVersions) == 0 {
			e.AcceptableTLSVersions = []string{"TLS 1.2", "TLS 1.3"}
		}
	}

	return nil
}
