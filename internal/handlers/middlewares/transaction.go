package middlewares

import (
	"fmt"
	"net/http"

	"github.com/Thospol/go-fiber/internal/core/context"
	"github.com/Thospol/go-fiber/internal/core/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// TransactionPostgresql to do transaction postgresql
func TransactionPostgresql(next http.Handler) fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		database := sql.PostgreDatabase.Begin()
		c.Locals(context.PostgreDatabaseKey, database)
		err = c.Next()
		if err != nil {
			_ = database.Rollback()
			return
		}

		if r := recover(); r != nil && r != http.ErrAbortHandler {
			_ = database.Rollback()

			var ok bool
			if err, ok = r.(error); !ok {
				// Set error that will call the global error handler
				err = fmt.Errorf("%v", r)
				logrus.Panic(err)
			}
		}

		if database.Commit().Error != nil {
			_ = database.Rollback()
		}

		return
	}
}

// TransactionMysql to do transaction mysql
func TransactionMysql(next http.Handler) fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		database := sql.MysqlDatabase.Begin()
		c.Locals(context.MysqlDatabaseKey, database)
		err = c.Next()
		if err != nil {
			_ = database.Rollback()
			return
		}

		if r := recover(); r != nil && r != http.ErrAbortHandler {
			_ = database.Rollback()

			var ok bool
			if err, ok = r.(error); !ok {
				// Set error that will call the global error handler
				err = fmt.Errorf("%v", r)
				logrus.Panic(err)
			}
		}

		if database.Commit().Error != nil {
			_ = database.Rollback()
		}

		return
	}
}
