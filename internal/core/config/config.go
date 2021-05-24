package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/fsnotify/fsnotify"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	// CF -> for use configs model
	CF = &Configs{}
)

// Environment environment
type Environment string

const (
	// LOCAL environment local
	LOCAL Environment = "local"
	// DEVELOP environment develop
	DEVELOP Environment = "dev"
	// PRODUCTION environment production
	PRODUCTION Environment = "prod"
)

// Develop environment develop
func (env Environment) Develop() bool {
	return env == DEVELOP
}

// Production environment production
func (env Environment) Production() bool {
	return env == PRODUCTION
}

// DatabaseConfig database config model
type DatabaseConfig struct {
	Host         string `mapstructure:"HOST"`
	Port         int    `mapstructure:"PORT"`
	Username     string `mapstructure:"USERNAME"`
	Password     string `mapstructure:"PASSWORD"`
	DatabaseName string `mapstructure:"DATABASE_NAME"`
	DriverName   string `mapstructure:"DRIVER_NAME"`
	Timeout      string `mapstructure:"TIMEOUT"`
	Enable       bool   `mapstructure:"ENABLE"`
}

// JWTExpireTimeConfig jwt expire time config model
type JWTExpireTimeConfig struct {
	Day    time.Duration `mapstructure:"DAY"`
	Hour   time.Duration `mapstructure:"HOUR"`
	Minute time.Duration `mapstructure:"MINUTE"`
}

// Configs config models
type Configs struct {
	UniversalTranslator *ut.UniversalTranslator
	Validator           *validator.Validate
	App                 struct {
		ProjectID   string `mapstructure:"PROJECT_ID"`
		WebBaseURL  string `mapstructure:"WEB_BASE_URL"`
		ApiBaseURL  string `mapstructure:"API_BASE_URL"`
		Release     bool   `mapstructure:"RELEASE"`
		Port        int    `mapstructure:"PORT"`
		Environment Environment
	} `mapstructure:"APP"`
	HTTPServer struct {
		ReadTimeout  time.Duration `mapstructure:"READ_TIMEOUT"`
		WriteTimeout time.Duration `mapstructure:"WRITE_TIMEOUT"`
		IdleTimeout  time.Duration `mapstructure:"IDLE_TIMEOUT"`
	} `mapstructure:"HTTP_SERVER"`
	SQL struct {
		PostgreSQL DatabaseConfig `mapstructure:"POSTGRE_SQL"`
		MySQL      DatabaseConfig `mapstructure:"MY_SQL"`
	} `mapstructure:"SQL"`
	Mongo   DatabaseConfig `mapstructure:"MONGO"`
	Redis   DatabaseConfig `mapstructure:"REDIS"`
	Swagger struct {
		Title       string   `mapstructure:"TITLE"`
		Version     string   `mapstructure:"VERSION"`
		Host        string   `mapstructure:"HOST"`
		BaseURL     string   `mapstructure:"BASE_URL"`
		Description string   `mapstructure:"DESCRIPTION"`
		Schemes     []string `mapstructure:"SCHEMES"`
		Enable      bool     `mapstructure:"ENABLE"`
	} `mapstructure:"SWAGGER"`
	JWT struct {
		SecretKey string `mapstructure:"SECRET_KEY"`
		Access    struct {
			JWTExpireTimeConfig `mapstructure:"EXPIRE_TIME"`
		} `mapstructure:"ACCESS"`
		Refresh struct {
			JWTExpireTimeConfig `mapstructure:"EXPIRE_TIME"`
		} `mapstructure:"REFRESH"`
	} `mapstructure:"JWT"`
}

// InitConfig init config
func InitConfig(configPath string, environment string) error {
	v := viper.New()
	CF.App.Environment = CF.parseEnvironment(environment)
	v.AddConfigPath(configPath)
	v.SetConfigName(fmt.Sprintf("config.%s", CF.App.Environment))
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetConfigType("yml")

	if err := v.ReadInConfig(); err != nil {
		logrus.Error("read config file error:", err)
		return err
	}

	if err := bindingConfig(v, CF); err != nil {
		logrus.Error("binding config error:", err)
		return err
	}

	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		if err := bindingConfig(v, CF); err != nil {
			logrus.Error("binding error:", err)
			return
		}
	})

	return nil
}

// bindingConfig binding config
func bindingConfig(vp *viper.Viper, cf *Configs) error {
	if err := vp.Unmarshal(&cf); err != nil {
		logrus.Error("unmarshal config error:", err)
		return err
	}

	validate := validator.New()

	if err := validate.RegisterValidation("maxString", validateString); err != nil {
		logrus.Error("cannot register maxString Validator config error:", err)
		return err
	}

	en := en.New()
	cf.UniversalTranslator = ut.New(en, en)
	enTrans, _ := cf.UniversalTranslator.GetTranslator("en")
	if err := en_translations.RegisterDefaultTranslations(validate, enTrans); err != nil {
		logrus.Error("cannot add english translator config error:", err)
		return err
	}
	_ = validate.RegisterTranslation("maxString", enTrans, func(ut ut.Translator) error {
		return ut.Add("maxString", "Sorry, {0} cannot exceed {1} characters", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		field := strings.ToLower(fe.Field())
		t, _ := ut.T("maxString", field, fe.Param())
		return t
	})

	cf.Validator = validate

	return nil
}

// validateString implements validator.Func for max string by rune
func validateString(fl validator.FieldLevel) bool {
	var err error

	limit := 255
	param := strings.Split(fl.Param(), `:`)
	if len(param) > 0 {
		limit, err = strconv.Atoi(param[0])
		if err != nil {
			limit = 255
		}
	}

	if lengthOfString := utf8.RuneCountInString(fl.Field().String()); lengthOfString > limit {
		return false
	}

	return true
}

func (c Configs) parseEnvironment(environment string) Environment {
	switch environment {
	case "local":
		return LOCAL

	case "dev":
		return DEVELOP

	case "prod":
		return PRODUCTION
	}

	return DEVELOP
}
