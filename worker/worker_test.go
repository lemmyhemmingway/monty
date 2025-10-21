package worker

import (
	"crypto/x509"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/monty/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}

	if err := db.AutoMigrate(&models.Endpoint{}, &models.Status{}, &models.SSLStatus{}); err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}

	return db
}

func TestParseHostPort(t *testing.T) {
	tests := []struct {
		input    string
		expected struct {
			host string
			port string
		}
	}{
		{"https://example.com", struct{ host, port string }{"example.com", "443"}},
		{"http://example.com", struct{ host, port string }{"example.com", "80"}},
		{"example.com:8443", struct{ host, port string }{"example.com", "8443"}},
		{"example.com", struct{ host, port string }{"example.com", "443"}},
	}

	for _, test := range tests {
		host, port, err := parseHostPort(test.input)
		if err != nil {
			t.Errorf("parseHostPort(%s) returned error: %v", test.input, err)
		}
		if host != test.expected.host || port != test.expected.port {
			t.Errorf("parseHostPort(%s) = (%s, %s), expected (%s, %s)",
				test.input, host, port, test.expected.host, test.expected.port)
		}
	}
}

func TestTLSVersionString(t *testing.T) {
	tests := []struct {
		version  uint16
		expected string
	}{
		{0x0301, "TLS 1.0"},
		{0x0302, "TLS 1.1"},
		{0x0303, "TLS 1.2"},
		{0x0304, "TLS 1.3"},
		{0x0000, "Unknown"},
	}

	for _, test := range tests {
		result := tlsVersionString(test.version)
		if result != test.expected {
			t.Errorf("tlsVersionString(%d) = %s, expected %s", test.version, result, test.expected)
		}
	}
}

func TestWorkerIsTLSVersionAcceptable(t *testing.T) {
	w := &Worker{}

	tests := []struct {
		version    string
		acceptable []string
		expected   bool
	}{
		{"TLS 1.2", []string{"TLS 1.2", "TLS 1.3"}, true},
		{"TLS 1.3", []string{"TLS 1.2", "TLS 1.3"}, true},
		{"TLS 1.1", []string{"TLS 1.2", "TLS 1.3"}, false},
		{"TLS 1.0", []string{"TLS 1.2", "TLS 1.3"}, false},
	}

	for _, test := range tests {
		result := w.isTLSVersionAcceptable(test.version, test.acceptable)
		if result != test.expected {
			t.Errorf("isTLSVersionAcceptable(%s, %v) = %v, expected %v",
				test.version, test.acceptable, result, test.expected)
		}
	}
}

func TestWorkerValidateCertificateChain(t *testing.T) {
	w := &Worker{}

	// This is a basic test since our implementation is simplified
	tests := []struct {
		certs    []*x509.Certificate
		expected bool
	}{
		{nil, false},
		{[]*x509.Certificate{}, false},
		{[]*x509.Certificate{{}, {}}, true}, // At least 2 certs
	}

	for _, test := range tests {
		result := w.validateCertificateChain(test.certs)
		if result != test.expected {
			t.Errorf("validateCertificateChain(%v) = %v, expected %v",
				test.certs, result, test.expected)
		}
	}
}

func TestWorkerSaveSSLStatus(t *testing.T) {
	db := setupTestDB(t)
	models.DB = db

	w := &Worker{}

	status := models.SSLStatus{
		ID:                   uuid.New().String(),
		EndpointID:           uuid.New().String(),
		CertificateExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		DaysUntilExpiry:      30,
		IsValid:              true,
		DomainMatches:        true,
		ChainValid:           true,
		Issuer:               "Test Issuer",
		Subject:              "Test Subject",
		TLSVersion:           "TLS 1.3",
		SerialNumber:         "12345",
		ErrorMessage:         "",
		CheckedAt:            time.Now(),
	}

	w.saveSSLStatus(status.EndpointID, status)

	// Verify it was saved
	var saved models.SSLStatus
	err := db.First(&saved, "id = ?", status.ID).Error
	if err != nil {
		t.Fatalf("failed to find saved SSL status: %v", err)
	}

	if saved.EndpointID != status.EndpointID {
		t.Errorf("saved EndpointID = %s, expected %s", saved.EndpointID, status.EndpointID)
	}
	if saved.IsValid != status.IsValid {
		t.Errorf("saved IsValid = %v, expected %v", saved.IsValid, status.IsValid)
	}
}
