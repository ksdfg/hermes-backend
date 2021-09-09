package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"hermes/config"
)

func Register(router fiber.Router) {
	// Get config vars
	cfg := config.Get()

	// Create new store and sessions map
	store := session.New()
	sessions := make(map[string]sessionData)

	// Register types
	store.RegisterType(sessionData{})
	store.RegisterType([]message{})

	// Create new service object
	svc := service{store: store, sessions: sessions, config: cfg}

	// Add handler at route
	router.Post("/", svc.new)
	router.Get("/loggedIn", svc.loggedIn)
	router.Post("/send", svc.send)
	router.Get("/logs", svc.logs)
}
