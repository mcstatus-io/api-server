package main

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func init() {
	app.Get("/ping", PingHandler)
	app.Use(NotFoundHandler)
}

// PingHandler responds with a 200 OK status for simple health checks.
func PingHandler(ctx *fiber.Ctx) error {
	return ctx.SendStatus(http.StatusOK)
}

// NotFoundHandler handles requests to routes that do not exist and returns a 404 Not Found status.
func NotFoundHandler(ctx *fiber.Ctx) error {
	return ctx.SendStatus(http.StatusNotFound)
}
