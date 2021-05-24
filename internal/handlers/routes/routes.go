package routes

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/Thospol/go-fiber/internal/core/config"
	"github.com/Thospol/go-fiber/internal/handlers"
	"github.com/Thospol/go-fiber/internal/handlers/middlewares"
	"github.com/Thospol/go-fiber/internal/pkg/user"

	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/sirupsen/logrus"
)

const (
	// MaximumSize100MB body limit 100 mb.
	MaximumSize100MB = 1024 * 1024 * 100
)

func NewRouter() {
	app := fiber.New(
		fiber.Config{
			IdleTimeout:  5 * time.Second,
			BodyLimit:    MaximumSize100MB,
			ErrorHandler: handlers.Errors("./public/500.html"),
		},
	)

	app.Use(
		compress.New(),
		requestid.New(),
		recover.New(),
		cors.New(),
	)

	app.Static("", "./public", fiber.Static{
		Compress: true,
	})

	api := app.Group("/api")
	v1 := api.Group("/v1")
	v1.Use(middlewares.AcceptLanguage())
	v1.Use(middlewares.Logger())
	if config.CF.Swagger.Enable {
		v1.Get("/swagger/*", swagger.Handler)
	}

	userEndpoint := user.NewEndpoint()
	users := v1.Group("users")
	users.Get("/:id", handlers.Cache(5*time.Second), userEndpoint.GetUser)

	api.Use(handlers.NotFound("./public/404.html"))

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		logrus.Info("Gracefully shutting down...")
		_ = app.Shutdown()
	}()

	logrus.Infof("Start server on port: %d ...", config.CF.App.Port)
	err := app.Listen(fmt.Sprintf(":%d", config.CF.App.Port))
	if err != nil {
		logrus.Panic(err)
	}
}
