package models

import (
	"time"
)

type DomainStatus struct {
	ID                    string    `gorm:"primaryKey" json:"id"`
	EndpointID            string    `gorm:"not null" json:"endpoint_id"`
	DomainExpiresAt       time.Time `json:"domain_expires_at"`
	DaysUntilExpiry       int       `json:"days_until_expiry"`
	IsRegistered          bool      `json:"is_registered"`
	Registrar             string    `json:"registrar"`
	ErrorMessage          string    `json:"error_message"`
	CheckedAt             time.Time `json:"checked_at"`
}

// TableName overrides the table name
func (DomainStatus) TableName() string {
	return "domain_statuses"
}
