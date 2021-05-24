package file

import (
	"fmt"
	"os"
	"sync"

	"github.com/Thospol/go-fiber/internal/core/config"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

const (
	fileKey = "file"
)

// Service user service interface
type Service interface {
	UploadFile(c *fiber.Ctx, database *gorm.DB) error
}

type service struct {
	config *config.Configs
	result *config.ReturnResult
	mux    sync.Mutex
}

// NewService new user service
func NewService() Service {
	return &service{
		config: config.CF,
		result: config.RR,
	}
}

// UploadFile upload file service
func (s *service) UploadFile(c *fiber.Ctx, database *gorm.DB) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	file, err := c.FormFile(fileKey)
	if err != nil {
		return err
	}

	dir := fmt.Sprintf("./%s", file.Filename)
	err = c.SaveFile(file, dir)
	if err != nil {
		return err
	}

	defer os.RemoveAll(dir)

	// TODO:

	return nil
}

// UploadFiles upload files service
func (s *service) UploadFiles(c *fiber.Ctx, database *gorm.DB) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	if form, err := c.MultipartForm(); err == nil {
		files := form.File[fileKey]

		for _, file := range files {
			dir := fmt.Sprintf("./%s", file.Filename)
			if err := c.SaveFile(file, dir); err != nil {
				return err
			}

			defer os.RemoveAll(dir)
		}
	}

	// TODO:

	return nil
}
