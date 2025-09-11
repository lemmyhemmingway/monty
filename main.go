package main

import (
	"github.com/gofiber/fiber/v2"
	"monty/handlers"
)

func main() {
	app := fiber.New()

	handlers.RegisterHealth(app)

	app.Listen(":3000")
}
