package handlers

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/monty/models"
)

func RegisterEndpoints(app *fiber.App) {
	app.Get("/endpoints", listEndpoints)
	app.Post("/endpoints", createEndpoint)
	app.Get("/endpoint-urls", listEndpointURLs)
}

func listEndpoints(c *fiber.Ctx) error {
	var endpoints []models.Endpoint
	models.DB.Find(&endpoints)
	return c.JSON(endpoints)
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
