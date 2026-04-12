package worker

import (
	"errors"

	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/engine/pkg/broker"
)

// handleRejectedTransition Обработать переход on.rejected для шагов типа action/compensate
func (r *Runner) handleRejectedTransition(
	event *broker.SagaStepResultEvent,
	sagaDef *domain.SagaDefinition,
	currentStepDef *domain.DefinitionStep,
	instance *domain.InstanceView,
	currentStep *domain.StepView,
) (result *eventHandleResult, err error) {
	return nil, errors.New("rejected transition is not supported yet")
}
