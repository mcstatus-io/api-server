package main

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func AuthenticateMiddleware() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		sessionToken := ctx.Get("Authorization")

		if len(sessionToken) < 1 {
			ctx.Locals("authUser", nil)

			return ctx.Next()
		}

		session, err := db.GetSessionByID(sessionToken)

		if err != nil {
			return err
		}

		if session == nil {
			return ctx.Status(http.StatusForbidden).SendString("Invalid or expired session")
		}

		user, err := db.GetUserByID(session.User)

		if err != nil {
			return err
		}

		if user == nil {
			ctx.Locals("authUser", nil)

			return ctx.Next()
		}

		ctx.Locals("authUser", user)

		return ctx.Next()
	}
}

func GetUserMiddleware(param string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userID := ctx.Params(param)

		if userID == "@me" {
			authUserLocal := ctx.Locals("authUser")

			if authUserLocal != nil {
				ctx.Locals("user", authUserLocal)

				return ctx.Next()
			}

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
			return ctx.Status(http.StatusNotFound).SendString("No user found by that ID")
		}

		ctx.Locals("user", user)

		return ctx.Next()
	}
}

func RequireAuthMiddleware() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		authUser, ok := ctx.Locals("authUser").(*User)

		if !ok || authUser == nil {
			return ctx.Status(http.StatusUnauthorized).SendString("You must be authorized to access this endpoint")
		}

		return ctx.Next()
	}
}

func UserAuthMiddleware() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		user, ok := ctx.Locals("user").(*User)

		if !ok || user == nil {
			return ctx.Status(http.StatusNotFound).SendString("User not found")
		}

		authUser, ok := ctx.Locals("authUser").(*User)

		if !ok || authUser == nil {
			return ctx.Status(http.StatusUnauthorized).SendString("You must be authorized to access this endpoint")
		}

		if authUser.ID == user.ID {
			return ctx.Next()
		}

		return ctx.Status(http.StatusForbidden).SendString("You must be authorized to access this endpoint")
	}
}
