package context

import (
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/Thospol/go-fiber/internal/core/config"
	"github.com/Thospol/go-fiber/internal/core/sql"
	"github.com/Thospol/go-fiber/internal/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

const (
	pathKey            = "path"
	compositeFormDepth = 3
	// UserKey user key
	UserKey = "user"
	// LangKey lang key
	LangKey = "lang"
	// PostgreDatabaseKey database `postgre` key
	PostgreDatabaseKey = "postgre_database"
	// PostgreDatabaseKey database `mysql` key
	MysqlDatabaseKey = "mysql_database"
	// UserKey parameters key
	ParametersKey = "parameters"
)

// Context custom fiber context
type Context interface {
	BindValue(i interface{}, validate bool) error
	GetPostgreDatabase() *gorm.DB
	GetMysqlDatabase() *gorm.DB
	GetUser() (*models.UserSession, error)
}

type context struct {
	*fiber.Ctx
}

// New new custom fiber context
func New(c *fiber.Ctx) Context {
	return &context{c}
}

// BindValue bind value
func (c *context) BindValue(i interface{}, validate bool) error {
	switch c.Method() {
	case http.MethodGet:
		_ = c.QueryParser(i)

	case http.MethodPost:
		_ = c.BodyParser(i)
	}

	c.PathParser(i, 1)
	c.Locals(ParametersKey, i)
	c.trimspace(i)

	if validate {
		err := c.validate(i)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetPostgreDatabase get connection database `postgresql`
func (c *context) GetPostgreDatabase() *gorm.DB {
	val := c.Locals(PostgreDatabaseKey)
	if val == nil {
		return sql.PostgreDatabase
	}

	return val.(*gorm.DB)
}

// GetMysqlDatabase get connection database `mysql`
func (c *context) GetMysqlDatabase() *gorm.DB {
	val := c.Locals(MysqlDatabaseKey)
	if val == nil {
		return sql.MysqlDatabase
	}

	return val.(*gorm.DB)
}

// GetUser get user session
func (c *context) GetUser() (*models.UserSession, error) {
	val := c.Locals(UserKey)
	if val == nil {
		return nil, config.RR.Internal.Unauthorized.WithLocale(c.Ctx)
	}

	return val.(*models.UserSession), nil
}

// PathParser parse path param
func (c *context) PathParser(i interface{}, depth int) {
	formValue := reflect.ValueOf(i)
	if formValue.Kind() == reflect.Ptr {
		formValue = formValue.Elem()
	}
	t := reflect.TypeOf(formValue.Interface())
	for i := 0; i < t.NumField(); i++ {
		fieldName := t.Field(i).Name
		paramValue := formValue.FieldByName(fieldName)
		if paramValue.IsValid() {
			if depth < compositeFormDepth && paramValue.Kind() == reflect.Struct {
				depth++
				c.PathParser(paramValue.Addr().Interface(), depth)
			}
			tag := t.Field(i).Tag.Get(pathKey)
			if tag != "" {
				setValue(paramValue, c.Params(tag))
			}
		}
	}
}

func setValue(paramValue reflect.Value, value string) {
	if paramValue.IsValid() && value != "" {
		switch paramValue.Kind() {
		case reflect.Uint:
			number, _ := strconv.ParseUint(value, 10, 32)
			paramValue.SetUint(number)

		case reflect.String:
			paramValue.SetString(value)

		default:
			number, err := strconv.Atoi(value)
			if err != nil {
				paramValue.SetString(value)
			} else {
				paramValue.SetInt(int64(number))
			}
		}
	}
}

func (c *context) validate(i interface{}) error {
	if err := config.CF.Validator.Struct(i); err != nil {
		return config.RR.CustomMessage(err.Error(), err.Error()).WithLocale(c.Ctx)
	}

	return nil
}

func (c *context) trimspace(i interface{}) {
	e := reflect.ValueOf(i).Elem()
	for i := 0; i < e.NumField(); i++ {
		if e.Type().Field(i).Type.Kind() != reflect.String {
			continue
		}

		value := e.Field(i).Interface().(string)
		e.Field(i).SetString(strings.TrimSpace(value))
	}
}
