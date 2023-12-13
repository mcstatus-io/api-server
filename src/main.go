package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var (
	app *fiber.App = fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			var fiberError *fiber.Error

			if errors.As(err, &fiberError) {
				return ctx.SendStatus(fiberError.Code)
			}

			log.Printf("Error: %v - URI: %s\n", err, ctx.Request().URI())

			return ctx.SendStatus(http.StatusInternalServerError)
		},
	})
	db         *MongoDB            = &MongoDB{}
	config     *Config             = DefaultConfig
	instanceID uint16              = 0
	validate   *validator.Validate = validator.New()
)

func init() {
	var err error

	if err = config.ReadFile("config.yml"); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			panic(err)
		}

		if err = config.WriteFile("config.yml"); err != nil {
			log.Fatalf("Failed to write config file: %v", err)
		}
	}

	if instanceID, err = GetInstanceID(); err != nil {
		panic(err)
	}

	if err := db.Connect(config.MongoDB); err != nil {
		panic(err)
	}

	log.Println("Successfully connected to MongoDB")

	app.Hooks().OnListen(func(ld fiber.ListenData) error {
		log.Printf("Listening on %s:%d\n", config.Host, config.Port+instanceID)

		return nil
	})
}

func main() {
	defer db.Close()

	if err := app.Listen(fmt.Sprintf("%s:%d", config.Host, config.Port+instanceID)); err != nil {
		panic(err)
	}
}
