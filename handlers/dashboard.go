package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monty/models"
)

func RegisterDashboard(app *fiber.App) {
	app.Get("/dashboard", showDashboard)
}

func showDashboard(c *fiber.Ctx) error {
	var endpoints []models.Endpoint
	models.DB.Find(&endpoints)

	var response []EndpointWithStatus
	for _, ep := range endpoints {
		status := calculateStatus(ep)
		response = append(response, EndpointWithStatus{
			Endpoint:     ep,
			StatusString: status,
		})
	}

	return c.Render("dashboard", fiber.Map{
		"endpoints": response,
	})
}
