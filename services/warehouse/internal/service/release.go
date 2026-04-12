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

type ReleaseResult struct {
	MovementIDs []uuid.UUID
}

// Release освобождает резервы в рамках одной транзакции с outbox
func (s *WarehouseService) Release(ctx context.Context, orderID uuid.UUID, sagaInstanceId uuid.UUID, stepId string) (*ReleaseResult, error) {
	var (
		result     *ReleaseResult
		releaseErr error
	)

	err := s.transactor.Transaction(ctx, func(ctx context.Context) error {
		reservations, err := s.repo.GetReservationsByOrderID(ctx, orderID)
		if err != nil {
			releaseErr = fmt.Errorf("get reservations: %w", err)
			return err
		}

		var movementIDs []uuid.UUID

		for _, reservation := range reservations {
			balance, err := s.repo.GetBalance(ctx, reservation.WarehouseID, reservation.ProductID)
			if err != nil {
				releaseErr = fmt.Errorf("get balance: %w", err)
				return err
			}
			if balance == nil {
				continue
			}

			balance.QuantityReserved -= reservation.Quantity
			if balance.QuantityReserved < 0 {
				balance.QuantityReserved = 0
			}
			if err := s.repo.UpdateBalanceWithLock(ctx, balance); err != nil {
				releaseErr = fmt.Errorf("update balance: %w", err)
				return err
			}

			movementID := uuid.New()
			movement := &domain.Movement{
				MovementID:         movementID,
				WarehouseID:        reservation.WarehouseID,
				ProductID:          reservation.ProductID,
				MovementType:       domain.MovementTypeRelease,
				Quantity:           reservation.Quantity,
				BusinessRef:        orderID.String(),
				ScenarioInstanceID: sagaInstanceId,
				StepID:             stepId,
			}
			if err := s.repo.CreateMovement(ctx, movement); err != nil {
				releaseErr = fmt.Errorf("create movement: %w", err)
				return err
			}
			movementIDs = append(movementIDs, movementID)

			if err := s.repo.UpdateReservationStatus(ctx, reservation.ID, domain.ReservationStatusCancelled); err != nil {
				releaseErr = fmt.Errorf("update reservation status: %w", err)
				return err
			}
		}

		result = &ReleaseResult{
			MovementIDs: movementIDs,
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
				"movement_ids": utils.MapSlice(movementIDs, func(u uuid.UUID) string {
					return u.String()
				}),
			},
		}
		if err = s.outboxRepo.PushSagaStepResultEvent(ctx, outboxResult); err != nil {
			releaseErr = fmt.Errorf("push outbox event: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, releaseErr
	}

	return result, nil
}
