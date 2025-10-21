package handlers

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/monty/models"
)

type EndpointWithUptime struct {
	models.Endpoint
	Uptime float64 `json:"uptime"`
}

func RegisterEndpoints(app *fiber.App) {
	app.Get("/endpoints", listEndpoints)
	app.Post("/endpoints", createEndpoint)
	app.Get("/endpoint-urls", listEndpointURLs)
	app.Get("/statuses", listStatuses)
	app.Get("/endpoints/:id/statuses", listEndpointStatuses)
	// SSL status endpoints
	app.Get("/ssl-statuses", listSSLStatuses)
	app.Get("/endpoints/:id/ssl-statuses", listEndpointSSLStatuses)
}

func calculateUptime(endpointID string) float64 {
	// Get endpoint config to determine check type
	var ep models.Endpoint
	if err := models.DB.First(&ep, "id = ?", endpointID).Error; err != nil {
		return 0
	}

	if ep.CheckType == "ssl" {
		return calculateSSLUptime(endpointID)
	}

	// HTTP uptime calculation
	var statuses []models.Status
	models.DB.Where("endpoint_id = ?", endpointID).Find(&statuses)

	if len(statuses) == 0 {
		return 0
	}

	expectedCodes := []int(ep.ExpectedStatusCodes)
	if len(expectedCodes) == 0 {
		expectedCodes = []int{200, 201, 202, 203, 204, 205, 206, 207, 208, 226, 300, 301, 302, 303, 304, 305, 307, 308}
	}

	successful := 0
	for _, status := range statuses {
		if status.ErrorMessage == "" { // No network error
			codeExpected := false
			for _, code := range expectedCodes {
				if status.Code == code {
					codeExpected = true
					break
				}
			}
			if codeExpected && status.ResponseTime <= ep.MaxResponseTime {
				successful++
			}
		}
	}

	return (float64(successful) / float64(len(statuses))) * 100
}

func calculateSSLUptime(endpointID string) float64 {
	var sslStatuses []models.SSLStatus
	models.DB.Where("endpoint_id = ?", endpointID).Find(&sslStatuses)

	if len(sslStatuses) == 0 {
		return 0
	}

	successful := 0
	for _, status := range sslStatuses {
		if status.IsValid {
			successful++
		}
	}

	return (float64(successful) / float64(len(sslStatuses))) * 100
}

func listEndpoints(c *fiber.Ctx) error {
	var endpoints []models.Endpoint
	models.DB.Find(&endpoints)

	var response []EndpointWithUptime
	for _, ep := range endpoints {
		uptime := calculateUptime(ep.ID)
		response = append(response, EndpointWithUptime{
			Endpoint: ep,
			Uptime:   uptime,
		})
	}

	return c.JSON(response)
}

func listEndpointURLs(c *fiber.Ctx) error {
	var endpoints []models.Endpoint
	models.DB.Find(&endpoints)
	urls := make([]string, len(endpoints))
	for i, ep := range endpoints {
		urls[i] = ep.URL
	}
	return c.JSON(urls)
}

func createEndpoint(c *fiber.Ctx) error {
	var input struct {
		URL                  string   `json:"url"`
		CheckType            string   `json:"check_type,omitempty"`            // optional, defaults to "http"
		Interval             int      `json:"interval"`
		Timeout              *int     `json:"timeout,omitempty"`               // optional, defaults to 30
		ExpectedStatusCodes  []int    `json:"expected_status_codes,omitempty"`  // optional, defaults to 2xx/3xx
		MaxResponseTime      *int     `json:"max_response_time,omitempty"`      // optional, defaults to 5000ms
		// SSL-specific fields
		MinDaysValid         *int     `json:"min_days_valid,omitempty"`         // optional, defaults to 30
		CheckChain           *bool    `json:"check_chain,omitempty"`            // optional, defaults to true
		CheckDomainMatch     *bool    `json:"check_domain_match,omitempty"`     // optional, defaults to true
		AcceptableTLSVersions []string `json:"acceptable_tls_versions,omitempty"` // optional, defaults to ["TLS 1.2", "TLS 1.3"]
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid input"})
	}

	input.URL = strings.TrimSpace(input.URL)
	if input.URL == "" || input.Interval <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "url and interval must be provided"})
	}

	// Set defaults for optional fields
	timeout := 30
	if input.Timeout != nil && *input.Timeout > 0 {
		timeout = *input.Timeout
	}

	maxResponseTime := 5000
	if input.MaxResponseTime != nil && *input.MaxResponseTime > 0 {
		maxResponseTime = *input.MaxResponseTime
	}

	// SSL-specific defaults
	checkType := "http"
	if input.CheckType != "" {
		checkType = input.CheckType
	}

	minDaysValid := 30
	if input.MinDaysValid != nil && *input.MinDaysValid > 0 {
		minDaysValid = *input.MinDaysValid
	}

	checkChain := true
	if input.CheckChain != nil {
		checkChain = *input.CheckChain
	}

	checkDomainMatch := true
	if input.CheckDomainMatch != nil {
		checkDomainMatch = *input.CheckDomainMatch
	}

	acceptableTLSVersions := models.StringArray{"TLS 1.2", "TLS 1.3"}
	if len(input.AcceptableTLSVersions) > 0 {
		acceptableTLSVersions = models.StringArray(input.AcceptableTLSVersions)
	}

	ep := models.Endpoint{
		ID:                   uuid.New().String(),
		URL:                  input.URL,
		CheckType:            checkType,
		Interval:             input.Interval,
		Timeout:              timeout,
		ExpectedStatusCodes:  models.IntArray(input.ExpectedStatusCodes),
		MaxResponseTime:      maxResponseTime,
		MinDaysValid:         minDaysValid,
		CheckChain:           checkChain,
		CheckDomainMatch:     checkDomainMatch,
		AcceptableTLSVersions: acceptableTLSVersions,
		CreatedAt:            time.Now(),
	}
	if err := models.DB.Create(&ep).Error; err != nil {
		if errors.Is(err, models.ErrInvalidEndpoint) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid endpoint configuration"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not create endpoint"})
	}
	return c.Status(fiber.StatusCreated).JSON(ep)
}

func listStatuses(c *fiber.Ctx) error {
	var statuses []models.Status
	models.DB.Order("checked_at desc").Find(&statuses)
	return c.JSON(statuses)
}

func listEndpointStatuses(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "endpoint id required"})
	}

	var statuses []models.Status
	models.DB.Where("endpoint_id = ?", id).Order("checked_at desc").Find(&statuses)
	return c.JSON(statuses)
}

func listSSLStatuses(c *fiber.Ctx) error {
	var sslStatuses []models.SSLStatus
	models.DB.Order("checked_at desc").Find(&sslStatuses)
	return c.JSON(sslStatuses)
}

func listEndpointSSLStatuses(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "endpoint id required"})
	}

	var sslStatuses []models.SSLStatus
	models.DB.Where("endpoint_id = ?", id).Order("checked_at desc").Find(&sslStatuses)
	return c.JSON(sslStatuses)
}
