package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

type AuthCallbackResponse struct {
	SessionToken string `json:"sessionToken"`
}

func init() {
	app.Get("/ping", PingHandler)
	app.Get("/users/@me", GetSelfUserHandler)
	app.Post("/auth/discord/callback", DiscordAuthCallback)
	app.Use(NotFoundHandler)
}

// PingHandler responds with a 200 OK status for simple health checks.
func PingHandler(ctx *fiber.Ctx) error {
	return ctx.SendStatus(http.StatusOK)
}

// GetSelfUserHandler returns the current user using the session token.
func GetSelfUserHandler(ctx *fiber.Ctx) error {
	sessionToken := ctx.Get("Authorization")

	if len(sessionToken) < 1 {
		return ctx.Status(http.StatusUnauthorized).SendString("Missing \"Authorization\" header")
	}

	session, err := m.GetSessionByID(sessionToken)

	if err != nil {
		log.Println(err)

		return ctx.SendStatus(http.StatusInternalServerError)
	}

	if session == nil {
		return ctx.SendStatus(http.StatusForbidden)
	}

	user, err := m.GetUserByID(session.User)

	if err != nil {
		log.Println(err)

		return ctx.SendStatus(http.StatusInternalServerError)
	}

	if user == nil {
		return ctx.SendStatus(http.StatusNotFound)
	}

	return ctx.JSON(user)
}

// DiscordAuthCallback is the authentication handler for a Discord login sequence.
func DiscordAuthCallback(ctx *fiber.Ctx) error {
	code := ctx.Query("code")

	if len(code) < 1 {
		return ctx.Status(http.StatusBadRequest).SendString("Missing \"code\" from query parameters")
	}

	exchangeResponse, err := ExchangeDiscordCode(code)

	if err != nil {
		log.Println(err)

		return ctx.Status(http.StatusInternalServerError).SendString("Failed to exchange code for Discord access token")
	}

	user, err := GetDiscordUser(exchangeResponse.AccessToken)

	if err != nil {
		log.Println(err)

		return ctx.Status(http.StatusInternalServerError).SendString("Failed to get Discord user with access token")
	}

	if err = m.UpsertUser(
		bson.M{"_id": user.ID},
		bson.M{
			"$setOnInsert": bson.M{
				"createdAt": time.Now(),
			},
			"$set": bson.M{
				"username": user.Username,
				"email":    user.Email,
				"avatar":   user.Avatar,
			},
		},
	); err != nil {
		log.Println(err)

		return ctx.SendStatus(http.StatusInternalServerError)
	}

	sessionToken, err := GenerateSessionToken()

	if err != nil {
		log.Println(err)

		return ctx.SendStatus(http.StatusInternalServerError)
	}

	if err = m.InsertSession(Session{
		ID:        sessionToken,
		User:      user.ID,
		CreatedAt: time.Now(),
	}); err != nil {
		log.Println(err)

		return ctx.SendStatus(http.StatusInternalServerError)
	}

	return ctx.JSON(AuthCallbackResponse{
		SessionToken: sessionToken,
	})
}

// NotFoundHandler handles requests to routes that do not exist and returns a 404 Not Found status.
func NotFoundHandler(ctx *fiber.Ctx) error {
	return ctx.SendStatus(http.StatusNotFound)
}
