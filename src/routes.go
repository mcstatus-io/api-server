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
	Email           string `json:"email" validate:"email,required"`
	Password        string `json:"password" validate:"min=6,required"`
	ConfirmPassword string `json:"confirmPassword" validate:"eqfield=Password,required"`
}

type PostApplicationsRequestBody struct {
	Name             string `json:"name" validate:"min=2,max=64,required"`
	ShortDescription string `json:"shortDescription" validate:"min=30,max=480,required"`
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
	app.Post("/auth/discord", PostDiscordCallbackHandler)
	app.Post("/auth/github", PostGitHubCallbackHandler)
	app.Get("/users/:id", AuthenticateMiddleware(), GetUserMiddleware("id"), UserAuthMiddleware(), GetUserHandler)
	app.Get("/users/:id/applications", AuthenticateMiddleware(), GetUserMiddleware("id"), UserAuthMiddleware(), GetUserApplicationsHandler)
	app.Post("/applications", AuthenticateMiddleware(), RequireAuthMiddleware(), PostApplicationsHandler)
	app.Get("/applications/:id", GetApplicationHandler)
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

	if user.Type != "local" {
		return ctx.Status(http.StatusForbidden).SendString("A user exists with that email but is not using local login. Please login with the other service provider instead.")
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

// PostDiscordCallbackHandler authenticates the user using the Discord OAuth code.
func PostDiscordCallbackHandler(ctx *fiber.Ctx) error {
	code := ctx.Query("code")

	if len(code) < 1 {
		return ctx.Status(http.StatusBadRequest).SendString("Missing code query parameter")
	}

	tokenResponse, err := ExchangeDiscordAccessToken(code)

	if err != nil {
		return err
	}

	discordUser, err := GetDiscordUser(tokenResponse.AccessToken)

	if err != nil {
		return err
	}

	user, err := db.GetUserByEmail(discordUser.Email)

	if err != nil {
		return err
	}

	var userID string

	if user == nil {
		userDocument := User{
			ID:        RandomHexString(8),
			Email:     discordUser.Email,
			Password:  tokenResponse.AccessToken,
			Type:      "discord",
			CreatedAt: time.Now().UTC(),
		}

		if err := db.InsertUser(userDocument); err != nil {
			return err
		}

		userID = userDocument.ID
	} else {
		if user.Type != "discord" {
			return ctx.Status(http.StatusForbidden).SendString("A user exists with that email but is not using Discord for login. Please login with the other service provider or local login instead.")
		}

		userID = user.ID
	}

	sessionDocument := Session{
		ID:        RandomHexString(16),
		User:      userID,
		CreatedAt: time.Now(),
	}

	if err := db.InsertSession(sessionDocument); err != nil {
		return err
	}

	return ctx.JSON(sessionDocument)
}

// PostGitHubCallbackHandler authenticates the user using the GitHub OAuth code.
func PostGitHubCallbackHandler(ctx *fiber.Ctx) error {
	code := ctx.Query("code")

	if len(code) < 1 {
		return ctx.Status(http.StatusBadRequest).SendString("Missing code query parameter")
	}

	tokenResponse, err := ExchangeGitHubAccessToken(code)

	if err != nil {
		return err
	}

	githubEmails, err := GetGitHubEmails(tokenResponse.AccessToken)

	if err != nil {
		return err
	}

	var primaryEmail string

	for _, emails := range githubEmails {
		if !emails.Primary {
			continue
		}

		primaryEmail = emails.Email
	}

	if len(primaryEmail) < 1 {
		return ctx.Status(http.StatusConflict).SendString("Cannot find a primary email address associated with that GitHub user")
	}

	user, err := db.GetUserByEmail(primaryEmail)

	if err != nil {
		return err
	}

	var userID string

	if user == nil {
		userDocument := User{
			ID:        RandomHexString(8),
			Email:     primaryEmail,
			Password:  tokenResponse.AccessToken,
			Type:      "github",
			CreatedAt: time.Now().UTC(),
		}

		if err := db.InsertUser(userDocument); err != nil {
			return err
		}

		userID = userDocument.ID
	} else {
		if user.Type != "github" {
			return ctx.Status(http.StatusForbidden).SendString("A user exists with that email but is not using Discord for login. Please login with the other service provider or local login instead.")
		}

		userID = user.ID
	}

	sessionDocument := Session{
		ID:        RandomHexString(16),
		User:      userID,
		CreatedAt: time.Now(),
	}

	if err := db.InsertSession(sessionDocument); err != nil {
		return err
	}

	return ctx.JSON(sessionDocument)
}

// GetUserHandler returns the user by the ID or the current authenticated user.
func GetUserHandler(ctx *fiber.Ctx) error {
	return ctx.JSON(ctx.Locals("user"))
}

// GetUserHandler returns the user by the ID or the current authenticated user.
func GetUserApplicationsHandler(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(*User)

	applications, err := db.GetApplicationsByUser(user.ID)

	if err != nil {
		return err
	}

	return ctx.JSON(applications)
}

// PostApplicationsHandler creates a new application using the body data provided.
func PostApplicationsHandler(ctx *fiber.Ctx) error {
	authUser := ctx.Locals("authUser").(*User)

	var requestBody PostApplicationsRequestBody

	if err := ctx.BodyParser(&requestBody); err != nil {
		return ctx.Status(http.StatusBadRequest).SendString(fmt.Sprintf("Invalid request body: %s", err))
	}

	applicationDocument := Application{
		ID:               RandomHexString(12),
		Name:             requestBody.Name,
		ShortDescription: requestBody.ShortDescription,
		User:             authUser.ID,
		Token:            RandomHexString(16),
		TotalRequests:    0,
		CreatedAt:        time.Now().UTC(),
	}

	if err := db.InsertApplication(applicationDocument); err != nil {
		return err
	}

	return ctx.Status(http.StatusCreated).JSON(applicationDocument)
}

// GetApplicationHandler returns the specific application by ID.
func GetApplicationHandler(ctx *fiber.Ctx) error {
	application, err := db.GetApplicationByID(ctx.Params("id"))

	if err != nil {
		return err
	}

	if application == nil {
		return ctx.Status(http.StatusNotFound).SendString("No application found by that ID")
	}

	return ctx.JSON(application)
}
