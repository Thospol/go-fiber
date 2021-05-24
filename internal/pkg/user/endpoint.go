package user

import (
	"github.com/Thospol/go-fiber/internal/core/config"
	"github.com/Thospol/go-fiber/internal/core/context"
	"github.com/Thospol/go-fiber/internal/core/render"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// Endpoint user endpoint interface
type Endpoint interface {
	GetUser(c *fiber.Ctx) error
}

type endpoint struct {
	config  *config.Configs
	result  *config.ReturnResult
	service Service
}

func NewEndpoint() Endpoint {
	return &endpoint{
		config:  config.CF,
		result:  config.RR,
		service: NewService(),
	}
}

// GetUser godoc
// @Tags User
// @Summary GetUser
// @Description Request get user by id
// @Accept json
// @Produce json
// @Param Accept-Language header string false "(en, th)" default(th)
// @Param id path string true "input id" default(1)
// @Success 200 {object} models.User
// @Failure 400 {object} config.SwaggerInfoResult
// @Failure 500 {object} config.SwaggerInfoResult
// @Security ApiKeyAuth
// @Router /users/{id} [get]
func (ep *endpoint) GetUser(c *fiber.Ctx) error {
	request := new(getUserRequest)
	ctx := context.New(c)
	err := ctx.BindValue(request, false)
	if err != nil {
		logrus.Errorf("[GetUser] bind value error: %s", err)
		return render.Error(c, err)
	}

	response, err := ep.service.GetUser(ctx.GetPostgreDatabase(), request.Id)
	if err != nil {
		logrus.Errorf("[GetUser] call service error: %s", err)
		return render.Error(c, err)
	}

	return render.JSON(c, response)
}
