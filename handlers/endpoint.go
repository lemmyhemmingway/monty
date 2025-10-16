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
}

func calculateUptime(endpointID string) float64 {
	var total int64
	var successful int64

	models.DB.Model(&models.Status{}).Where("endpoint_id = ?", endpointID).Count(&total)
	models.DB.Model(&models.Status{}).Where("endpoint_id = ? AND code >= 200 AND code < 300", endpointID).Count(&successful)

	if total == 0 {
		return 0
	}
	return (float64(successful) / float64(total)) * 100
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
		URL      string `json:"url"`
		Interval int    `json:"interval"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid input"})
	}

	input.URL = strings.TrimSpace(input.URL)
	if input.URL == "" || input.Interval <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "url and interval must be provided"})
	}

	ep := models.Endpoint{
		ID:        uuid.New().String(),
		URL:       input.URL,
		Interval:  input.Interval,
		CreatedAt: time.Now(),
	}
	if err := models.DB.Create(&ep).Error; err != nil {
		if errors.Is(err, models.ErrInvalidEndpoint) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "url and interval must be provided"})
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
