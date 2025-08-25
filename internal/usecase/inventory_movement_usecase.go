package usecase

import (
	"context"
	"errors"
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
	"tokobahankue/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type InventoryMovementUseCase struct {
	DB                          *gorm.DB
	Log                         *logrus.Logger
	Validate                    *validator.Validate
	InventoryMovementRepository *repository.InventoryMovementRepository
	BranchInventoryRepository   *repository.BranchInventoryRepository
}

func NewInventoryMovementUseCase(db *gorm.DB, logger *logrus.Logger, validate *validator.Validate,
	inventoryMovementRepository *repository.InventoryMovementRepository, branchInventoryRepository *repository.BranchInventoryRepository) *InventoryMovementUseCase {
	return &InventoryMovementUseCase{
		DB:                          db,
		Log:                         logger,
		Validate:                    validate,
		InventoryMovementRepository: inventoryMovementRepository,
		BranchInventoryRepository:   branchInventoryRepository,
	}
}

func (c *InventoryMovementUseCase) Create(ctx context.Context, request *model.BulkCreateInventoryMovementRequest) (*model.BulkInventoryMovementResponse, error) {

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return nil, errors.New("bad request")
	}

	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	var responses []model.InventoryMovementResponse

	for _, m := range request.Movements {
		branchInv := &entity.BranchInventory{}
		err := c.BranchInventoryRepository.FindByBranchIDAndSizeID(tx, branchInv, request.BranchID, m.SizeID)

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Belum ada → buat baru
			branchInv = &entity.BranchInventory{
				BranchID: request.BranchID,
				SizeID:   m.SizeID,
				Stock:    m.ChangeQty,
			}
			if err := c.BranchInventoryRepository.Create(tx, branchInv); err != nil {
				return nil, errors.New("error creating branch_inventory")
			}
		} else if err != nil {
			c.Log.WithError(err).Error("error querying branch_inventory")
			return nil, errors.New("internal server error")
		} else {
			// Sudah ada → update stok
			if err := c.BranchInventoryRepository.UpdateStock(tx, branchInv.ID, m.ChangeQty); err != nil {
				return nil, errors.New("error updating stock")
			}
		}

		// Catat movement
		movement := &entity.InventoryMovement{
			BranchInventoryID: branchInv.ID,
			ChangeQty:         m.ChangeQty,
			ReferenceType:     request.ReferenceType,
			ReferenceKey:      request.ReferenceKey,
		}
		if err := c.InventoryMovementRepository.Create(tx, movement); err != nil {
			return nil, errors.New("error creating inventory_movement")
		}

		responses = append(responses, model.InventoryMovementResponse{
			ID:        movement.ID,
			ChangeQty: movement.ChangeQty,
			CreatedAt: movement.CreatedAt,
		})
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("commit transaction failed")
	}

	return &model.BulkInventoryMovementResponse{
		ReferenceType: request.ReferenceType,
		ReferenceKey:  request.ReferenceKey,
		Movements:     responses,
	}, nil
}
