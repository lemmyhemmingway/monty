package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/monty/models"
)

func RegisterDashboard(app *fiber.App) {
	app.Get("/dashboard", showDashboard)
}

func showDashboard(c *fiber.Ctx) error {
	var endpoints []models.Endpoint
	models.DB.Find(&endpoints)

	// Group endpoints by type
	grouped := make(map[string][]EndpointWithStatus)
	for i, ep := range endpoints {
		status := calculateStatus(ep)
		item := EndpointWithStatus{
			Endpoint:     ep,
			StatusString: status,
		}
		// Add sequential ID
		item.Endpoint.ID = fmt.Sprintf("%d", i+1)
		grouped[ep.CheckType] = append(grouped[ep.CheckType], item)
	}

	return c.Render("dashboard", fiber.Map{
		"groupedEndpoints": grouped,
	})
}
