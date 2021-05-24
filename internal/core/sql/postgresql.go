package sql

import (
	"fmt"

	"github.com/Thospol/go-fiber/internal/core/config"

	"gorm.io/driver/postgres"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	// PostgreDatabase global variable database `postgresql`
	PostgreDatabase = &gorm.DB{}
)

// InitConnectionPostgreSQL open initialize a new db connection.
func InitConnectionPostgreSQL(config config.DatabaseConfig) (err error) {
	postgreSQLCredentials := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		config.Host,
		config.Port,
		config.Username,
		config.Password,
		config.DatabaseName,
	)

	PostgreDatabase, err = gorm.Open(postgres.Open(postgreSQLCredentials), &gorm.Config{
		PrepareStmt: true,
	})
	if err != nil {
		logrus.Errorf("[InitConnectionPostgresqlSQL] failed to connect to the database error: %s", err)
		return err
	}

	sqlDB, err := PostgreDatabase.DB()
	if err != nil {
		logrus.Errorf("[InitConnectionPostgresqlSQL] set up to connect to the database error: %s", err)
		return err
	}

	err = sqlDB.Ping()
	if err != nil {
		logrus.Errorf("[InitConnectionPostgresqlSQL] ping database error: %s", err)
		return err
	}

	return nil
}
