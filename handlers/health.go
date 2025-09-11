package handlers

import "github.com/gofiber/fiber/v2"

func RegisterHealth(app *fiber.App) {
	app.Get("/health", healthHandler)
}

func healthHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}
