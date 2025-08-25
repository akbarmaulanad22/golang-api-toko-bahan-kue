package repository

import (
	"tokobahankue/internal/entity"

	"github.com/sirupsen/logrus"
)

type InventoryMovementRepository struct {
	Repository[entity.InventoryMovement]
	Log *logrus.Logger
}

func NewInventoryMovementRepository(log *logrus.Logger) *InventoryMovementRepository {
	return &InventoryMovementRepository{
		Log: log,
	}
}
