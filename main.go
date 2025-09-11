package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monty/handlers"
	"github.com/monty/worker"
	"time"
)

func main() {
	app := fiber.New()

	handlers.RegisterHealth(app)

	worker.Start([]worker.Endpoint{
		{URL: "http://localhost:3000/health", Interval: 10 * time.Second},
	})

	app.Listen(":3000")
}
