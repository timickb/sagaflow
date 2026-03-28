package broker

import "context"

// StepResultHandler - вызывается для каждого прочитанного события
type StepResultHandler func(ctx context.Context, event *SagaStepResultEvent) error

type StepResultReader interface {
	Start(ctx context.Context, handler StepResultHandler) error
	Stop() error
}
