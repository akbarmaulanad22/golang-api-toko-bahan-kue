package usecase

import (
	"context"
	"errors"
	"strings"
	"time"
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
	"tokobahankue/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SaleDetailUseCase struct {
	DB                            *gorm.DB
	Log                           *logrus.Logger
	Validate                      *validator.Validate
	SaleDetailRepository          *repository.SaleDetailRepository
	SaleRepository                *repository.SaleRepository
	InventoryMovementRepository   *repository.InventoryMovementRepository
	BranchInventoryRepository     *repository.BranchInventoryRepository
	CashBankTransactionRepository *repository.CashBankTransactionRepository
}

func NewSaleDetailUseCase(
	db *gorm.DB,
	logger *logrus.Logger,
	validate *validator.Validate,
	saleDetailRepository *repository.SaleDetailRepository,
	saleRepository *repository.SaleRepository,
	inventoryMovementRepository *repository.InventoryMovementRepository,
	branchInventoryRepository *repository.BranchInventoryRepository,
	cashBankTransactionRepository *repository.CashBankTransactionRepository,
) *SaleDetailUseCase {
	return &SaleDetailUseCase{
		DB:                            db,
		Log:                           logger,
		Validate:                      validate,
		SaleDetailRepository:          saleDetailRepository,
		SaleRepository:                saleRepository,
		InventoryMovementRepository:   inventoryMovementRepository,
		BranchInventoryRepository:     branchInventoryRepository,
		CashBankTransactionRepository: cashBankTransactionRepository,
	}
}

func (c *SaleDetailUseCase) Cancel(ctx context.Context, request *model.CancelSaleDetailRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("error validating request body")
		return errors.New("bad request")
	}

	sale := new(entity.Sale)
	if err := c.SaleRepository.FindLockByCode(tx, request.SaleCode, sale); err != nil {
		c.Log.WithError(err).Error("error getting sale")
		return errors.New("not found")
	}

	createdTime := time.UnixMilli(sale.CreatedAt)
	now := time.Now()

	// Hitung durasi sejak dibuat
	duration := now.Sub(createdTime)

	// Jika status BUKAN PENDING dan sudah lewat 24 jam => tolak
	if duration.Hours() >= 24 {
		c.Log.WithField("sale_code", sale.Code).Error("error updating sale: exceeded 24-hour window")
		return errors.New("forbidden")
	}

	// ambil price dan qty detail
	detail := new(entity.SaleDetail)
	if err := c.SaleDetailRepository.FindPriceBySizeIDAndSaleCode(tx, sale.Code, request.SizeID, detail); err != nil {
		c.Log.WithError(err).Error("error updating sale detail")
		return errors.New("internal server error")
	}

	if detail.IsCancelled {
		c.Log.Error("error updating sale detail")
		return errors.New("forbidden")
	}

	// Lanjut update status
	if err := c.SaleDetailRepository.CancelBySizeID(tx, sale.Code, request.SizeID); err != nil {
		c.Log.WithError(err).Error("error updating sale detail")
		return errors.New("internal server error")
	}

	branchInv := &entity.BranchInventory{}
	if err := c.BranchInventoryRepository.FindByBranchIDAndSizeID(tx, branchInv, request.BranchID, request.SizeID); err != nil {
		c.Log.WithError(err).Error("error querying branch inventory")
		return errors.New("internal server error")
	}

	if err := c.BranchInventoryRepository.UpdateStock(tx, branchInv.ID, detail.Qty); err != nil {
		return errors.New("error updating stock")
	}

	// Catat movement
	movement := &entity.InventoryMovement{
		BranchInventoryID: branchInv.ID,
		ChangeQty:         detail.Qty,
		ReferenceType:     "SALE CANCELLED",
		ReferenceKey:      request.SaleCode,
	}

	if err := c.InventoryMovementRepository.Create(tx, movement); err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			switch mysqlErr.Number {
			case 1452:
				if strings.Contains(mysqlErr.Message, "FOREIGN KEY (`branch_inventory_id`)") {
					c.Log.Warn("branch inventory doesnt exists")
					return errors.New("invalid branch inventory id")
				}
				return errors.New("foreign key constraint failed")
			}
		}
		return errors.New("error creating inventory movement")
	}

	// penyesuaian total price sale dengan data detail
	sale.TotalPrice -= (detail.SellPrice * float64(detail.Qty))

	cashBankTransaction := entity.CashBankTransaction{
		TransactionDate: time.Now().UnixMilli(),
		Type:            "OUT",
		Source:          "SALE",
		Amount:          detail.SellPrice * float64(detail.Qty),
		Description:     "PENJUALAN PER ITEM DIBATALKAN",
		ReferenceKey:    sale.Code,
		BranchID:        &sale.BranchID,
	}

	if err := c.CashBankTransactionRepository.Create(tx, &cashBankTransaction); err != nil {
		return err
	}

	if err := c.SaleRepository.UpdateTotalPrice(tx, sale.Code, sale.TotalPrice); err != nil {
		c.Log.WithError(err).Error("error updating sale")
		return errors.New("internal server error")
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("error updating sale detail")
		return errors.New("internal server error")
	}

	return nil
}
