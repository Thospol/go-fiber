package user

import (
	"time"

	"github.com/Thospol/go-fiber/internal/core/config"
	"github.com/Thospol/go-fiber/internal/models"

	"gorm.io/gorm"
)

// Service user service interface
type Service interface {
	GetUser(database *gorm.DB, userId uint) (*models.User, error)
}

type service struct {
	config *config.Configs
	result *config.ReturnResult
}

// NewService new user service
func NewService() Service {
	return &service{
		config: config.CF,
		result: config.RR,
	}
}

// GetUser get user
func (s *service) GetUser(database *gorm.DB, userId uint) (*models.User, error) {
	return &models.User{
		Model: models.Model{
			ID:        userId,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Pronoun: "Mr.",
		Name:    "Thosapol",
	}, nil
}
