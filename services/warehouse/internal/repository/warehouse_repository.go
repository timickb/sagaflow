package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timickb/sagaflow/lib/db"
	"github.com/timickb/sagaflow/services/warehouse/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WarehouseRepository struct {
	db *db.Database
}

func NewWarehouseRepository(db *db.Database) *WarehouseRepository {
	return &WarehouseRepository{db: db}
}

func (r *WarehouseRepository) GetProductBySKU(ctx context.Context, sku string) (*domain.Product, error) {
	var product domain.Product
	err := r.db.WithTxSupport(ctx).Where("sku = ?", sku).First(&product).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

func (r *WarehouseRepository) GetProductByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	var product domain.Product
	err := r.db.WithTxSupport(ctx).Where("id = ?", id).First(&product).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

func (r *WarehouseRepository) GetBalance(ctx context.Context, warehouseID, productID uuid.UUID) (*domain.Balance, error) {
	var balance domain.Balance
	err := r.db.WithTxSupport(ctx).
		Where("warehouse_id = ? AND product_id = ?", warehouseID, productID).
		First(&balance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &balance, nil
}

func (r *WarehouseRepository) GetDefaultWarehouse(ctx context.Context) (*domain.Warehouse, error) {
	var warehouse domain.Warehouse
	err := r.db.WithTxSupport(ctx).First(&warehouse).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &warehouse, nil
}

func (r *WarehouseRepository) UpdateBalanceWithLock(ctx context.Context, balance *domain.Balance) error {
	result := r.db.WithTxSupport(ctx).Table("balances").
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("warehouse_id = ? AND product_id = ? AND version = ?",
			balance.WarehouseID, balance.ProductID, balance.Version).
		Updates(map[string]interface{}{
			"quantity_reserved": balance.QuantityReserved,
			"version":           balance.Version + 1,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("optimistic lock failed: balance was modified")
	}

	return nil
}

func (r *WarehouseRepository) CreateMovement(ctx context.Context, movement *domain.Movement) error {
	return r.db.WithTxSupport(ctx).Create(movement).Error
}

func (r *WarehouseRepository) CreateReservation(ctx context.Context, reservation *domain.Reservation) error {
	return r.db.WithTxSupport(ctx).Create(reservation).Error
}

func (r *WarehouseRepository) GetReservationsByOrderID(ctx context.Context, orderID uuid.UUID) ([]domain.Reservation, error) {
	var reservations []domain.Reservation
	err := r.db.WithTxSupport(ctx).
		Where("order_id = ? AND status = ?", orderID, domain.ReservationStatusActive).
		Find(&reservations).Error
	if err != nil {
		return nil, err
	}
	return reservations, nil
}

func (r *WarehouseRepository) UpdateReservationStatus(ctx context.Context, id uuid.UUID, status domain.ReservationStatus) error {
	return r.db.WithTxSupport(ctx).
		Model(&domain.Reservation{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *WarehouseRepository) GetWarehouseByID(ctx context.Context, id uuid.UUID) (*domain.Warehouse, error) {
	var warehouse domain.Warehouse
	err := r.db.WithTxSupport(ctx).Where("id = ?", id).First(&warehouse).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &warehouse, nil
}
