package models

import "time"

type SSLStatus struct {
	ID                   string    `gorm:"primaryKey" json:"id"`
	EndpointID           string    `json:"endpoint_id"`
	CertificateExpiresAt time.Time `json:"certificate_expires_at"`
	DaysUntilExpiry      int       `json:"days_until_expiry"`
	IsValid              bool      `json:"is_valid"`
	DomainMatches        bool      `json:"domain_matches"`
	ChainValid           bool      `json:"chain_valid"`
	Issuer               string    `json:"issuer"`
	Subject              string    `json:"subject"`
	TLSVersion           string    `json:"tls_version"`
	SerialNumber         string    `json:"serial_number"`
	ErrorMessage         string    `json:"error_message"`
	CheckedAt            time.Time `json:"checked_at"`
}
