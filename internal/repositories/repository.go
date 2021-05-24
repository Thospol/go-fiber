package repositories

import (
	"github.com/Thospol/go-fiber/internal/models"

	"gorm.io/gorm"
)

// Repository repository interface
type Repository interface {
	Create(database *gorm.DB, i interface{}) error
	Update(database *gorm.DB, i interface{}) error
	Delete(database *gorm.DB, i interface{}) error
	FindByID(database *gorm.DB, id uint, i interface{}) error
	BulkInsert(database *gorm.DB, sliceValue interface{}) error
}

type repository struct{}

// NewRepository new repository
func NewRepository() Repository {
	return &repository{}
}

// Create create record database
func (repo *repository) Create(database *gorm.DB, i interface{}) error {
	if m, ok := i.(models.ModelInterface); ok {
		m.Stamp()
	}

	if err := database.Create(i).Error; err != nil {
		return err
	}

	return nil
}

// Update update record database
func (repo *repository) Update(database *gorm.DB, i interface{}) error {
	if m, ok := i.(models.ModelInterface); ok {
		m.UpdateStamp()
	}

	if err := database.Save(i).Error; err != nil {
		return err
	}

	return nil
}

// Delete delete record database
func (repo *repository) Delete(database *gorm.DB, i interface{}) error {
	if m, ok := i.(models.ModelInterface); ok {
		m.DeleteStamp()
	}

	if err := database.Delete(i).Error; err != nil {
		return err
	}

	return nil
}

// FindByID find by id record database
func (repo *repository) FindByID(database *gorm.DB, id uint, i interface{}) error {
	if err := database.First(i, id).Error; err != nil {
		return err
	}

	return nil
}

// BulkInsert bulk insert into database
func (repo *repository) BulkInsert(database *gorm.DB, sliceValue interface{}) error {
	if result := database.Create(sliceValue); result.Error != nil {
		return result.Error
	}

	return nil
}
