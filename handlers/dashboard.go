package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/monty/models"
)

func RegisterDashboard(app *fiber.App) {
	app.Get("/", serveReactApp)
	app.Get("/endpoints", serveReactApp)
	app.Get("/old-dashboard", showDashboard) // Keep old HTML dashboard for reference
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

func serveReactApp(c *fiber.Ctx) error {
	// For development, serve the React dev server
	// For production, serve the built index.html
	htmlContent := `<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <link rel="icon" type="image/svg+xml" href="/vite.svg" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Monty Dashboard</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/static/assets/index.js"></script>
  </body>
</html>`
	return c.Type("html").SendString(htmlContent)
}
