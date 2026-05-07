package instance

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/lib/utils"
)

func (u *Usecase) Start(ctx context.Context, dto *domain.InstanceStartDto) (instanceId uuid.UUID, err error) {
	now := time.Now()
	sagaDef, sagaFound := u.cache.GetSagaDefinition(domain.SagaDefinitionHeader{
		Name:    dto.SagaName,
		Version: dto.SagaVersion,
	})
	if !sagaFound {
		return uuid.Nil, errors.New("saga not found")
	}

	startStepDef, ok := sagaDef.StepById[sagaDef.StartStepId]
	if !ok {
		return uuid.Nil, errors.New("start step not declared in saga")
	}
	dto.StartStepName = sagaDef.StartStepId

	inputDataRaw := make(map[string]any)
	for _, inputDef := range startStepDef.Inputs {
		if inputDef.SourceNamespace != domain.StepInputSourceInputContext {
			return uuid.Nil, errors.New("start step input data could refer only to initial context")
		}
		data, findErr := dto.InitialContext.Find(inputDef.SourcePath)
		if findErr != nil {
			return uuid.Nil, fmt.Errorf("find path in initial context: %w", findErr)
		}
		inputDataRaw[inputDef.DestinationParam] = data
	}
	inputData, err := domain.NewJsonInstanceContextFromAny(inputDataRaw)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create JsonInstanceContext for input data: %w", err)
	}

	if startStepDef.Delay != nil {
		if dto.StartAfter == nil {
			dto.StartAfter = utils.Ptr(now.Add(*startStepDef.Delay))
		} else {
			dto.StartAfter = utils.Ptr(dto.StartAfter.Add(*startStepDef.Delay))
		}
	}

	err = u.transactor.Transaction(ctx, func(ctx context.Context) error {
		instanceId, err = u.repo.Create(ctx, dto)
		if err != nil {
			return fmt.Errorf("failed to save instance: %w", err)
		}
		_, err = u.stepRepo.Create(ctx, &domain.StepCreateDto{
			InstanceId: instanceId,
			StepName:   sagaDef.StartStepId,
			StepOrder:  1,
			InputData:  inputData,
		})
		if err != nil {
			return fmt.Errorf("failed to save start step: %w", err)
		}

		return nil
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to start instance: %w", err)
	}

	// TODO: отправлять аналитику
	return instanceId, nil
}
