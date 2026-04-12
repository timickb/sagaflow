package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timickb/sagaflow/lib/broker"
	"github.com/timickb/sagaflow/lib/utils"
	"github.com/timickb/sagaflow/services/warehouse/internal/domain"
)

type ReserveItem struct {
	SKU      string
	Quantity int
}

type ReserveResult struct {
	ReservationID uuid.UUID
	MovementIDs   []uuid.UUID
}

// Reserve выполняет резервирование товаров в рамках одной транзакции с outbox
func (s *WarehouseService) Reserve(ctx context.Context, orderID uuid.UUID, sagaInstanceId uuid.UUID, stepId string, items []ReserveItem) (*ReserveResult, error) {
	var (
		result     *ReserveResult
		reserveErr error
	)

	err := s.transactor.Transaction(ctx, func(ctx context.Context) error {
		warehouse, err := s.repo.GetDefaultWarehouse(ctx)
		if err != nil {
			reserveErr = fmt.Errorf("get default warehouse: %w", err)
			return err
		}
		if warehouse == nil {
			reserveErr = fmt.Errorf("no warehouse found")
			return fmt.Errorf("no warehouse found")
		}

		var movementIDs []uuid.UUID
		reservationID := uuid.New()

		for _, item := range items {
			product, err := s.repo.GetProductBySKU(ctx, item.SKU)
			if err != nil {
				reserveErr = fmt.Errorf("get product by sku %s: %w", item.SKU, err)
				return err
			}
			if product == nil {
				reserveErr = fmt.Errorf("product not found: %s", item.SKU)
				return fmt.Errorf("product not found: %s", item.SKU)
			}

			balance, err := s.repo.GetBalance(ctx, warehouse.ID, product.ID)
			if err != nil {
				reserveErr = fmt.Errorf("get balance: %w", err)
				return err
			}
			if balance == nil {
				reserveErr = fmt.Errorf("no balance for product %s on warehouse %s", item.SKU, warehouse.Code)
				return fmt.Errorf("no balance for product %s on warehouse %s", item.SKU, warehouse.Code)
			}

			available := balance.QuantityTotal - balance.QuantityReserved
			if available < item.Quantity {
				reserveErr = fmt.Errorf("insufficient stock for %s: available=%d, requested=%d",
					item.SKU, available, item.Quantity)
				return fmt.Errorf("insufficient stock")
			}

			balance.QuantityReserved += item.Quantity
			if err = s.repo.UpdateBalanceWithLock(ctx, balance); err != nil {
				reserveErr = fmt.Errorf("update balance: %w", err)
				return err
			}

			movementID := uuid.New()
			movement := &domain.Movement{
				MovementID:         movementID,
				WarehouseID:        warehouse.ID,
				ProductID:          product.ID,
				MovementType:       domain.MovementTypeReserve,
				Quantity:           item.Quantity,
				BusinessRef:        orderID.String(),
				ScenarioInstanceID: sagaInstanceId,
				StepID:             stepId,
			}
			if err = s.repo.CreateMovement(ctx, movement); err != nil {
				reserveErr = fmt.Errorf("create movement: %w", err)
				return err
			}
			movementIDs = append(movementIDs, movementID)

			reservation := &domain.Reservation{
				ID:                 uuid.New(),
				OrderID:            orderID,
				WarehouseID:        warehouse.ID,
				ProductID:          product.ID,
				Quantity:           item.Quantity,
				Status:             domain.ReservationStatusActive,
				ScenarioInstanceID: sagaInstanceId,
			}
			if err = s.repo.CreateReservation(ctx, reservation); err != nil {
				reserveErr = fmt.Errorf("create reservation: %w", err)
				return err
			}
		}

		result = &ReserveResult{
			ReservationID: reservationID,
			MovementIDs:   movementIDs,
		}

		outboxResult := &broker.SagaStepResultEvent{
			Ref: broker.SagaStepRef{
				SagaId:      sagaInstanceId,
				StepName:    stepId,
				ServiceName: serviceName,
			},
			Worker: broker.WorkerInfo{
				InstanceId: "", // TODO
				Hostname:   "",
			},
			Status:     broker.SagaStepStatusCommitted,
			ResolvedAt: utils.Ptr(time.Now()),
			Result: map[string]any{
				"reservation_id": reservationID.String(),
				"movement_ids": utils.MapSlice(movementIDs, func(u uuid.UUID) string {
					return u.String()
				}),
			},
		}

		if err = s.outboxRepo.PushSagaStepResultEvent(ctx, outboxResult); err != nil {
			reserveErr = fmt.Errorf("push outbox event: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, reserveErr
	}

	return result, nil
}
