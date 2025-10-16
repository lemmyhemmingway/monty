package main

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/monty/handlers"
	"github.com/monty/models"
	"github.com/monty/worker"
)

func main() {
	models.ConnectDatabase()
	if err := models.Seed(); err != nil {
		panic(err)
	}

	app := fiber.New()

	handlers.RegisterHealth(app)
	handlers.RegisterEndpoints(app)

	var eps []models.Endpoint
	models.DB.Find(&eps)
	var workerEps []worker.Endpoint
	for _, ep := range eps {
		workerEps = append(workerEps, worker.Endpoint{ID: ep.ID, URL: ep.URL, Interval: time.Duration(ep.Interval) * time.Second})
	}
	w := worker.NewWorker(1 * time.Minute) // Check for new endpoints every minute
	w.Start(workerEps)

	app.Listen(":3000")
}
