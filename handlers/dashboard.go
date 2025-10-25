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

	var response []EndpointWithUptime
	for _, ep := range endpoints {
		uptime := calculateUptime(ep.ID)
		response = append(response, EndpointWithUptime{
			Endpoint: ep,
			Uptime:   uptime,
		})
	}

	return c.Render("dashboard", fiber.Map{
		"endpoints": response,
	})
}
