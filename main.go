package main

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/monty/handlers"
	"github.com/monty/models"
	"github.com/monty/worker"
)

func main() {
	models.ConnectDatabase()
	if err := models.Seed(); err != nil {
		panic(err)
	}

	engine := html.New("./templates", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	handlers.RegisterHealth(app)
	handlers.RegisterEndpoints(app)
	handlers.RegisterDashboard(app)

	var eps []models.Endpoint
	models.DB.Find(&eps)
	var workerEps []worker.Endpoint
	for _, ep := range eps {
		workerEps = append(workerEps, worker.Endpoint{
		ID:                  ep.ID,
		URL:                 ep.URL,
		Interval:            time.Duration(ep.Interval) * time.Second,
		Timeout:             time.Duration(ep.Timeout) * time.Second,
		ExpectedStatusCodes: []int(ep.ExpectedStatusCodes),
		MaxResponseTime:     time.Duration(ep.MaxResponseTime) * time.Millisecond,
	})
	}
	w := worker.NewWorker(1 * time.Minute) // Check for new endpoints every minute
	w.Start(workerEps)

	app.Listen(":3000")
}
