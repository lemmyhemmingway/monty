package main

import (
	"embed"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/monty/handlers"
	"github.com/monty/models"
	"github.com/monty/worker"
)

//go:embed static/*
var staticFiles embed.FS

func main() {
	models.ConnectDatabase()
	if err := models.Seed(); err != nil {
		panic(err)
	}

	app := fiber.New()

	app.Use(cors.New())

	// Serve static files from embedded FS in production
	app.Static("/static", "./static") // Fallback for development
	app.Get("/static/*", func(c *fiber.Ctx) error {
		path := c.Params("*")
		data, err := staticFiles.ReadFile("static/" + path)
		if err != nil {
			return c.Status(404).SendString("File not found")
		}
		c.Set("Content-Type", getContentType(path))
		return c.Send(data)
	})

	api := app.Group("/api")
	handlers.RegisterHealth(api)
	handlers.RegisterEndpoints(api)

	// Serve React app for all other routes
	app.Get("/*", func(c *fiber.Ctx) error {
		return c.SendFile("./static/index.html")
	})

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
	// Start server in a goroutine
	go func() {
		app.Listen(":3000")
	}()

	// Start worker after a short delay to ensure server is up
	time.Sleep(1 * time.Second)
	worker.StartGlobalWorker(workerEps)

	// Wait forever
	select {}
}

func getContentType(filename string) string {
	if strings.HasSuffix(filename, ".js") {
		return "application/javascript"
	}
	if strings.HasSuffix(filename, ".css") {
		return "text/css"
	}
	return "text/plain"
}
