package api

import (
	"hermes/config"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func Register(router fiber.Router) {
	// Get config vars
	cfg := config.Get()

	// Create session map
	sessions := make(map[uuid.UUID]sessionData)

	// Create new service object
	svc := service{session: sessions, config: cfg}

	// Add handlers at routes

	router.Post("/", svc.new)
	router.Post("/send", svc.send)

	router.Get("/loggedIn", svc.loggedIn)
	router.Get("/logs", svc.logs)

	router.Delete("/", svc.cleanup)
}
