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

// Confirm подтверждает резерв
func (s *WarehouseService) Confirm(ctx context.Context, orderID uuid.UUID, sagaInstanceId uuid.UUID, stepId string) error {
	var confirmErr error

	err := s.transactor.Transaction(ctx, func(ctx context.Context) error {
		reservations, err := s.repo.GetReservationsByOrderID(ctx, orderID)
		if err != nil {
			confirmErr = fmt.Errorf("get reservations: %w", err)
			return err
		}

		for _, reservation := range reservations {
			if err = s.repo.UpdateReservationStatus(ctx, reservation.ID, domain.ReservationStatusConfirmed); err != nil {
				confirmErr = fmt.Errorf("update reservation status: %w", err)
				return err
			}
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
			Result:     map[string]any{},
		}
		if err = s.outboxRepo.PushSagaStepResultEvent(ctx, outboxResult); err != nil {
			confirmErr = fmt.Errorf("push outbox event: %w", err)
		}
		return nil
	})

	if err != nil {
		return confirmErr
	}
	return nil
}
