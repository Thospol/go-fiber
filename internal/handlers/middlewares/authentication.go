package middlewares

import (
	"strings"

	"github.com/Thospol/go-fiber/internal/core/config"
	"github.com/Thospol/go-fiber/internal/core/context"
	"github.com/Thospol/go-fiber/internal/core/jwt"
	"github.com/Thospol/go-fiber/internal/core/redis"
	"github.com/Thospol/go-fiber/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/sirupsen/logrus"
)

const (
	authHeader        = "Authorization"
	prefixHeaderValue = "Bearer"
)

// RequireAuthentication require authentication
func RequireAuthentication() fiber.Handler {
	return func(c *fiber.Ctx) error {
		store := session.New()
		sess, err := store.Get(c)
		if err != nil {
			panic(err)
		}

		defer func() {
			_ = sess.Destroy()
		}()

		claims, err := verifyToken(c)
		if err != nil {
			logrus.Error("[RequireAuthentication] verify token error: ", config.RR.Internal.Unauthorized.Error())
			return c.
				Status(config.RR.Internal.Unauthorized.HTTPStatusCode()).
				JSON(config.RR.Internal.Unauthorized.WithLocale(c))
		}

		user, err := extractTokenMetadata(claims)
		if err != nil || redis.GetConnection().Get(user.AccessUUID, &user.Id) != nil {
			logrus.Error("[RequireAuthentication] extract token metadata error: ", config.RR.Internal.Unauthorized.Error())
			return c.
				Status(config.RR.Internal.Unauthorized.HTTPStatusCode()).
				JSON(config.RR.Internal.Unauthorized.WithLocale(c))
		}

		// Add the user session to locals
		c.Locals(context.UserKey, user)
		return c.Next()
	}
}

func extractToken(c *fiber.Ctx) string {
	token := strings.Replace(c.Get(authHeader), prefixHeaderValue, "", 1)
	return token
}

func verifyToken(c *fiber.Ctx) (map[string]interface{}, error) {
	claims, err := jwt.Parsed(extractToken(c), true)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

func extractTokenMetadata(claims map[string]interface{}) (*models.UserSession, error) {
	userId, _ := claims["sub"].(uint)
	accessUUID, _ := claims["access_uuid"].(string)
	refreshUUID, _ := claims["refresh_uuid"].(string)
	userSession := &models.UserSession{
		Id:          userId,
		AccessUUID:  accessUUID,
		RefreshUUID: refreshUUID,
	}

	return userSession, nil
}
