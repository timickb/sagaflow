package instance

import (
	"context"

	"github.com/timickb/sagaflow/engine/pkg/broker"
)

// ApplyStepResult - применить результат, отправленный обработчиком
func (u *Usecase) ApplyStepResult(ctx context.Context, event *broker.SagaStepResultEvent) error {
	// TODO: implement
	panic("implement me")
}
