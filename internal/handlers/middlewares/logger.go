package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/Thospol/go-fiber/internal/core/config"
	"github.com/Thospol/go-fiber/internal/core/context"
	"github.com/Thospol/go-fiber/internal/core/utils"
	"github.com/Thospol/go-fiber/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// Logger is log request
func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		if err != nil {
			return err
		}

		// write response log
		logs := logrus.Fields{
			"host":            c.Hostname(),
			"method":          c.Method(),
			"path":            c.OriginalURL(),
			"Accept-Language": c.Locals(context.LangKey),
			"clientIP":        c.IP(),
			"User-Agent":      c.Get("User-Agent"),
			"body-size":       fmt.Sprintf("%.5f MB", float64(bytes.NewReader(c.Request().Body()).Len())/1024.00/1024.00),
			"statusCode":      fmt.Sprintf("%d %s", c.Response().StatusCode(), http.StatusText(c.Response().StatusCode())),
			"processTime":     time.Since(start),
		}

		val := c.Locals(context.UserKey)
		if user, ok := val.(*models.UserSession); ok {
			logs["user_id"] = user.Id
		}

		if parameters := c.Locals(context.ParametersKey); parameters != nil {
			formValue := reflect.ValueOf(parameters)
			if formValue.Kind() == reflect.Ptr {
				formValue = formValue.Elem()
			}

			for i := 0; i < formValue.NumField(); i++ {
				valueField := formValue.Field(i)
				if fieldName := formValue.Type().Field(i).Name; isAboutPassword(fieldName) {
					paramValue := formValue.FieldByName(fieldName)
					password, ok := valueField.Interface().(string)
					if ok {
						paramValue.Set(reflect.ValueOf(utils.WrapPassword(password)))
					}
				}
			}
			parametersByte, _ := json.Marshal(parameters)
			logs["parameters"] = string(parametersByte)
		}

		if !strings.HasPrefix(c.OriginalURL(), fmt.Sprintf("%s/swagger", config.CF.Swagger.BaseURL)) {
			logrus.WithFields(logs).Infof("[%s][%s] response: %v", c.Method(), c.OriginalURL(), string(c.Response().Body()))
		}

		return nil
	}
}

func isAboutPassword(fieldName string) bool {
	return fieldName == "Password" ||
		fieldName == "CurrentPassword" ||
		fieldName == "NewPassword" ||
		fieldName == "Pin"
}
