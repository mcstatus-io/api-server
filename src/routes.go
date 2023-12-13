package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type PostLoginRequestBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type PostSignupRequestBody struct {
	FirstName       string `json:"firstName" validate:"min=1,max=36,required"`
	Email           string `json:"email" validate:"email,required"`
	Password        string `json:"password" validate:"min=6,required"`
	ConfirmPassword string `json:"confirmPassword" validate:"eqfield=Password,required"`
}

func init() {
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	/* app.Use(favicon.New(favicon.Config{
		Data: assets.Favicon,
	})) */

	if config.Environment == "development" {
		app.Use(cors.New(cors.Config{
			AllowOrigins:  "*",
			AllowMethods:  "HEAD,OPTIONS,GET,POST",
			ExposeHeaders: "X-Cache-Hit,X-Cache-Time-Remaining",
		}))

		app.Use(logger.New(logger.Config{
			Format:     "${time} ${ip}:${port} -> ${status}: ${method} ${path} (${latency})\n",
			TimeFormat: "2006/01/02 15:04:05",
		}))
	}

	app.Get("/ping", PingHandler)
	app.Post("/auth/login", PostLoginHandler)
	app.Post("/auth/signup", PostSignupHandler)
	app.Get("/users/:id", GetUserHandler)
}

// PingHandler responds with a 200 OK status for simple health checks.
func PingHandler(ctx *fiber.Ctx) error {
	return ctx.SendStatus(http.StatusOK)
}

// PostLoginHander authenticates the user with the login information they provide, creating a session.
func PostLoginHandler(ctx *fiber.Ctx) error {
	var requestBody PostLoginRequestBody

	if err := ctx.BodyParser(&requestBody); err != nil {
		return ctx.Status(http.StatusBadRequest).SendString(fmt.Sprintf("Invalid request body: %s", err))
	}

	user, err := db.GetUserByEmail(requestBody.Email)

	if err != nil {
		return err
	}

	if user == nil {
		return ctx.Status(http.StatusForbidden).SendString("No user exists with that email address")
	}

	if HashPassword(requestBody.Password) != user.Password {
		return ctx.Status(http.StatusForbidden).SendString("Invalid password")
	}

	sessionDocument := Session{
		ID:        RandomHexString(16),
		User:      user.ID,
		CreatedAt: time.Now(),
	}

	if err := db.InsertSession(sessionDocument); err != nil {
		return err
	}

	return ctx.JSON(sessionDocument)
}

// PostSignupHandler creates a new user with the information, and returns a new session.
func PostSignupHandler(ctx *fiber.Ctx) error {
	var requestBody PostSignupRequestBody

	if err := ctx.BodyParser(&requestBody); err != nil {
		return ctx.Status(http.StatusBadRequest).SendString(err.Error())
	}

	if err := validate.Struct(requestBody); err != nil {
		return ctx.Status(http.StatusBadRequest).SendString(err.Error())
	}

	existingUser, err := db.GetUserByEmail(requestBody.Email)

	if err != nil {
		return err
	}

	if existingUser != nil {
		return ctx.Status(http.StatusConflict).SendString("A user already exists with that username")
	}

	userDocument := User{
		ID:        RandomHexString(8),
		FirstName: requestBody.FirstName,
		Email:     requestBody.Email,
		Password:  HashPassword(requestBody.Password),
		CreatedAt: time.Now(),
	}

	if err := db.InsertUser(userDocument); err != nil {
		return err
	}

	sessionDocument := Session{
		ID:        RandomHexString(16),
		User:      userDocument.ID,
		CreatedAt: time.Now(),
	}

	if err := db.InsertSession(sessionDocument); err != nil {
		return err
	}

	return ctx.Status(http.StatusCreated).JSON(sessionDocument)
}

// GetUserHandler returns the user by the ID or the current authenticated user.
func GetUserHandler(ctx *fiber.Ctx) error {
	userID := ctx.Params("id")

	if userID == "@me" {
		sessionToken := ctx.Get("Authorization")

		if len(sessionToken) < 1 {
			return ctx.Status(http.StatusUnauthorized).SendString("Missing Authorization header")
		}

		session, err := db.GetSessionByID(sessionToken)

		if err != nil {
			return err
		}

		if session == nil {
			return ctx.Status(http.StatusForbidden).SendString("Invalid or expired session")
		}

		userID = session.User
	}

	user, err := db.GetUserByID(userID)

	if err != nil {
		return err
	}

	if user == nil {
		return ctx.Status(http.StatusNotFound).SendString("User not found by that ID")
	}

	return ctx.JSON(user)
}
