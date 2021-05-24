package sql

import (
	"fmt"

	"github.com/Thospol/go-fiber/internal/core/config"
	"github.com/sirupsen/logrus"

	"gorm.io/driver/mysql"

	"gorm.io/gorm"
)

var (
	// Database global variable database
	MysqlDatabase = &gorm.DB{}
)

// InitConnectionMysql open initialize a new db connection.
func InitConnectionMysql(config config.DatabaseConfig) (err error) {
	dns := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.DatabaseName,
	)

	MysqlDatabase, err = gorm.Open(mysql.Open(dns), &gorm.Config{})
	if err != nil {
		return err
	}

	sqlDB, err := MysqlDatabase.DB()
	if err != nil {
		logrus.Errorf("[InitConnectionMysql] set up to connect to the database error: %s", err)
		return err
	}

	err = sqlDB.Ping()
	if err != nil {
		logrus.Errorf("[InitConnectionMysql] ping database error: %s", err)
		return err
	}

	return nil
}
