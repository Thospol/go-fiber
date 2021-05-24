package handlers

import (
	"github.com/gofiber/fiber/v2/middleware/session"
)

var (
	store = session.New()
)

func NewSession() {

}
