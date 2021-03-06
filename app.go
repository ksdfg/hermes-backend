package main

import (
	"flag"
	"fmt"
	"log"

	"hermes/api"
	"hermes/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func init() {
	// Set logging format
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Read config variables
	config.Init()
}

var (
	prod = flag.Bool("prod", false, "Enable prefork in Production")
)

func main() {
	// Fetch config vars
	cfg := config.Get()

	// Parse command-line flags
	flag.Parse()

	// Create fiber app
	app := fiber.New(fiber.Config{
		Prefork: *prod, // go run app.go -prod
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{AllowOrigins: cfg.AllowOrigins, ExposeHeaders: "session"}))
	app.Use(compress.New())

	// Register handlers
	api.Register(app)

	// Listen on port 3000
	log.Fatal(app.Listen(fmt.Sprintf(":%d", cfg.Port))) // go run app.go -port=:3000
}
