package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/monty/models"
	"gorm.io/driver/sqlite"
"gorm.io/gorm"
)

func setupTestDB(t *testing.T) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}

	if err := db.AutoMigrate(&models.Endpoint{}, &models.Status{}); err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}

	models.DB = db

	sqlDB, err := db.DB()
	if err == nil {
		t.Cleanup(func() {
			sqlDB.Close()
		})
	}
}

func newTestApp(t *testing.T) *fiber.App {
	t.Helper()
	setupTestDB(t)

	app := fiber.New()
	RegisterEndpoints(app)
	return app
}

func TestCreateEndpointSuccess(t *testing.T) {
	app := newTestApp(t)

	payload := `{"url":"http://example.com","interval":5}`
	req := httptest.NewRequest(http.MethodPost, "/endpoints", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("failed to perform request: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	var body models.Endpoint
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body.URL != "http://example.com" {
		t.Fatalf("expected URL to match payload, got %s", body.URL)
	}
	if body.Interval != 5 {
		t.Fatalf("expected interval 5, got %d", body.Interval)
	}
	if body.ID == "" {
		t.Fatalf("expected generated ID to be set")
	}

	var stored models.Endpoint
	if err := models.DB.First(&stored, "id = ?", body.ID).Error; err != nil {
		t.Fatalf("expected endpoint persisted: %v", err)
	}
}

func TestCreateEndpointValidation(t *testing.T) {
	app := newTestApp(t)

	payload := `{"url":"   ","interval":0}`
	req := httptest.NewRequest(http.MethodPost, "/endpoints", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("failed to perform request: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestListEndpoints(t *testing.T) {
	app := newTestApp(t)

	eps := []models.Endpoint{
		{ID: uuid.New().String(), URL: "http://service-a", Interval: 10},
		{ID: uuid.New().String(), URL: "http://service-b", Interval: 20},
	}
	if err := models.DB.Create(&eps).Error; err != nil {
		t.Fatalf("failed to seed endpoints: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/endpoints", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("failed to perform request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var body []EndpointWithUptime
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(body) != 2 {
		t.Fatalf("expected 2 endpoints, got %d", len(body))
	}

	// Check that uptime is calculated (should be 0 since no statuses exist yet)
	for _, ep := range body {
		if ep.Uptime != 0 {
			t.Fatalf("expected uptime 0 for endpoint with no statuses, got %f", ep.Uptime)
		}
	}
}

func TestListEndpointURLs(t *testing.T) {
	app := newTestApp(t)

	eps := []models.Endpoint{
		{ID: uuid.New().String(), URL: "http://service-a", Interval: 10},
		{ID: uuid.New().String(), URL: "http://service-b", Interval: 20},
	}
	if err := models.DB.Create(&eps).Error; err != nil {
		t.Fatalf("failed to seed endpoints: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/endpoint-urls", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("failed to perform request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var urls []string
	if err := json.NewDecoder(resp.Body).Decode(&urls); err != nil {
		t.Fatalf("failed to decode url list: %v", err)
	}

	if len(urls) != 2 {
		t.Fatalf("expected 2 urls, got %d", len(urls))
	}

	expected := map[string]struct{}{"http://service-a": {}, "http://service-b": {}}
	for _, url := range urls {
		if _, ok := expected[url]; !ok {
			t.Fatalf("unexpected url in list: %s", url)
		}
	}
}

func TestListEndpointsWithUptime(t *testing.T) {
	app := newTestApp(t)

	epID := uuid.New().String()
	ep := models.Endpoint{
		ID: epID, URL: "http://service-a", Interval: 10,
	}
	if err := models.DB.Create(&ep).Error; err != nil {
		t.Fatalf("failed to seed endpoint: %v", err)
	}

	// Create some statuses: 3 successful (200), 1 failed (500), 1 error (0)
	statuses := []models.Status{
		{ID: uuid.New().String(), EndpointID: epID, Code: 200, CheckedAt: time.Now()},
		{ID: uuid.New().String(), EndpointID: epID, Code: 200, CheckedAt: time.Now()},
		{ID: uuid.New().String(), EndpointID: epID, Code: 200, CheckedAt: time.Now()},
		{ID: uuid.New().String(), EndpointID: epID, Code: 500, CheckedAt: time.Now()},
		{ID: uuid.New().String(), EndpointID: epID, Code: 0, CheckedAt: time.Now()},
	}
	if err := models.DB.Create(&statuses).Error; err != nil {
		t.Fatalf("failed to seed statuses: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/endpoints", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("failed to perform request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var body []EndpointWithUptime
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(body) != 1 {
		t.Fatalf("expected 1 endpoint, got %d", len(body))
	}

	// Expected uptime: 3 successful out of 5 total = 60%
	expectedUptime := 60.0
	if body[0].Uptime != expectedUptime {
		t.Fatalf("expected uptime %f, got %f", expectedUptime, body[0].Uptime)
	}
}

func TestListStatuses(t *testing.T) {
	app := newTestApp(t)

	// Create test endpoints
	ep1ID := uuid.New().String()
	ep2ID := uuid.New().String()
	eps := []models.Endpoint{
		{ID: ep1ID, URL: "http://service-a", Interval: 10},
		{ID: ep2ID, URL: "http://service-b", Interval: 20},
	}
	if err := models.DB.Create(&eps).Error; err != nil {
		t.Fatalf("failed to seed endpoints: %v", err)
	}

	// Create test statuses
	statuses := []models.Status{
		{ID: uuid.New().String(), EndpointID: ep1ID, Code: 200, CheckedAt: time.Now().Add(-time.Minute)},
		{ID: uuid.New().String(), EndpointID: ep2ID, Code: 500, CheckedAt: time.Now()},
	}
	if err := models.DB.Create(&statuses).Error; err != nil {
		t.Fatalf("failed to seed statuses: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/statuses", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("failed to perform request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var body []models.Status
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(body) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(body))
	}

	// Check that they're ordered by checked_at desc (most recent first)
	if body[0].Code != 500 || body[1].Code != 200 {
		t.Fatalf("expected statuses ordered by checked_at desc, got codes: %d, %d", body[0].Code, body[1].Code)
	}
}

func TestListEndpointStatuses(t *testing.T) {
	app := newTestApp(t)

	// Create test endpoints
	ep1ID := uuid.New().String()
	ep2ID := uuid.New().String()
	eps := []models.Endpoint{
		{ID: ep1ID, URL: "http://service-a", Interval: 10},
		{ID: ep2ID, URL: "http://service-b", Interval: 20},
	}
	if err := models.DB.Create(&eps).Error; err != nil {
		t.Fatalf("failed to seed endpoints: %v", err)
	}

	// Create test statuses - some for ep1, some for ep2
	statuses := []models.Status{
		{ID: uuid.New().String(), EndpointID: ep1ID, Code: 200, CheckedAt: time.Now().Add(-time.Minute)},
		{ID: uuid.New().String(), EndpointID: ep1ID, Code: 404, CheckedAt: time.Now()},
		{ID: uuid.New().String(), EndpointID: ep2ID, Code: 500, CheckedAt: time.Now()},
	}
	if err := models.DB.Create(&statuses).Error; err != nil {
		t.Fatalf("failed to seed statuses: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/endpoints/"+ep1ID+"/statuses", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("failed to perform request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var body []models.Status
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(body) != 2 {
		t.Fatalf("expected 2 statuses for endpoint, got %d", len(body))
	}

	// Check that all returned statuses belong to the correct endpoint
	for _, status := range body {
		if status.EndpointID != ep1ID {
			t.Fatalf("expected all statuses to belong to endpoint %s, got %s", ep1ID, status.EndpointID)
		}
	}
}
