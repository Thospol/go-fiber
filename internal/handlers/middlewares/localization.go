package middlewares

import (
	"github.com/Thospol/go-fiber/internal/core/context"

	"github.com/gofiber/fiber/v2"
)

const (
	// EN english language
	EN = "en"

	// TH thai language
	TH = "th"
)

// AcceptLanguage header Accept-Language
func AcceptLanguage() fiber.Handler {
	return func(c *fiber.Ctx) error {
		lang := c.Get("Accept-Language")
		if lang != EN && lang != TH {
			lang = EN
		}

		// Add the language to locals
		c.Locals(context.LangKey, lang)
		return c.Next()
	}
}
