package main

import (
	"embed"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/template/html/v2"
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

	engine := html.New("./templates", ".html")
	engine.AddFunc("contains", strings.Contains)
	engine.AddFunc("add", func(a, b int) int { return a + b })
	engine.AddFunc("inc", func(a int) int { return a + 1 })
	app := fiber.New(fiber.Config{
		Views: engine,
	})

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
	// Start server in a goroutine
	go func() {
		app.Listen(":3000")
	}()

	// Start worker after a short delay to ensure server is up
	time.Sleep(1 * time.Second)
	w := worker.NewWorker(1 * time.Minute) // Check for new endpoints every minute
	w.Start(workerEps)

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
