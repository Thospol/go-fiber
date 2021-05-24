package main

import (
	"flag"
	"fmt"

	"github.com/Thospol/go-fiber/docs"
	"github.com/Thospol/go-fiber/internal/core/config"
	"github.com/Thospol/go-fiber/internal/core/jwt"
	"github.com/Thospol/go-fiber/internal/core/mongodb"
	"github.com/Thospol/go-fiber/internal/core/redis"
	"github.com/Thospol/go-fiber/internal/core/sql"
	"github.com/Thospol/go-fiber/internal/handlers/routes"

	stackdriver "github.com/TV4/logrus-stackdriver-formatter"
	"github.com/sirupsen/logrus"
)

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	environment := flag.String("environment", "local", "set working environment")
	configs := flag.String("config", "configs", "set configs path, default as: 'configs'")

	flag.Parse()

	// Init configuration
	err := config.InitConfig(*configs, *environment)
	if err != nil {
		panic(err)
	}
	//=======================================================

	// programatically set swagger info
	docs.SwaggerInfo.Title = config.CF.Swagger.Title
	docs.SwaggerInfo.Description = config.CF.Swagger.Description
	docs.SwaggerInfo.Version = config.CF.Swagger.Version
	docs.SwaggerInfo.Host = fmt.Sprintf("%s%s", config.CF.Swagger.Host, config.CF.Swagger.BaseURL)
	//=======================================================

	// set logrus
	if config.CF.App.Release {
		logrus.SetFormatter(stackdriver.NewFormatter(
			stackdriver.WithService("api"),
			stackdriver.WithVersion("v1.0.0")))
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{})
	}
	logrus.Infof("Initial 'Configuration'. %+v", config.CF)
	//=======================================================

	// Init return result
	err = config.InitReturnResult("configs")
	if err != nil {
		panic(err)
	}
	//=======================================================

	// Get SecretKey JWT
	jwt.LoadKey()
	// =======================================================

	// Init connection postgresql
	if config.CF.SQL.PostgreSQL.Enable {
		err = sql.InitConnectionPostgreSQL(config.CF.SQL.PostgreSQL)
		if err != nil {
			panic(err)
		}
	}
	//========================================================

	// Init connection mysql
	if config.CF.SQL.MySQL.Enable {
		err = sql.InitConnectionMysql(config.CF.SQL.MySQL)
		if err != nil {
			panic(err)
		}
	}
	//========================================================

	// Init connection mongoDB
	if config.CF.Mongo.Enable {
		err = mongodb.InitDatabase(&mongodb.Options{
			URL:          config.CF.Mongo.Host,
			Port:         config.CF.Mongo.Port,
			Username:     config.CF.Mongo.Username,
			Password:     config.CF.Mongo.Password,
			DatabaseName: config.CF.Mongo.DatabaseName,
			Debug:        !config.CF.App.Release,
		})
		if err != nil {
			panic(err)
		}
	}
	// =======================================================

	// Init connection redis
	if config.CF.Redis.Enable {
		conf := redis.Configuration{
			Host:     config.CF.Redis.Host,
			Port:     config.CF.Redis.Port,
			Password: config.CF.Redis.Password,
		}
		if err := redis.Init(conf); err != nil {
			panic(err)
		}
	}
	//========================================================

	// New router
	routes.NewRouter()
	//========================================================
}
