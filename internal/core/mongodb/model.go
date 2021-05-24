package mongodb

import (
	"time"

	"github.com/Thospol/go-fiber/internal/core/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Model common mongodb model
type Model struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt" bson:"created_at,omitempty"`
	UpdatedAt *time.Time         `json:"updatedAt,omitempty" bson:"updated_at,omitempty"`
	DeletedAt *time.Time         `json:"deletedAt,omitempty" bson:"deleted_at,omitempty"`
}

// ModelInterface model interface
type ModelInterface interface {
	GetID() primitive.ObjectID
	SetID(id primitive.ObjectID)
	Stamp()
	UpdateStamp()
	DeleteStamp()
	GetCreatedAt() time.Time
}

// SetID set id
func (model *Model) SetID(id primitive.ObjectID) {
	model.ID = id
}

// GetID get id
func (model *Model) GetID() primitive.ObjectID {
	return model.ID
}

// Stamp current time to model
func (model *Model) Stamp() {
	timeNow := utils.NowWhichNonZeroMilliseconds()
	model.UpdatedAt = &timeNow
	model.CreatedAt = timeNow
}

// UpdateStamp current updated at model
func (model *Model) UpdateStamp() {
	timeNow := utils.NowWhichNonZeroMilliseconds()
	model.UpdatedAt = &timeNow
}

// DeleteStamp current deleted at model
func (model *Model) DeleteStamp() {
	timeNow := utils.NowWhichNonZeroMilliseconds()
	model.DeletedAt = &timeNow
}

// GetCreatedAt get created_at
func (model *Model) GetCreatedAt() time.Time {
	return model.CreatedAt
}
