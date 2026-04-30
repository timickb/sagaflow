package worker

import (
	"context"

	"github.com/timickb/sagaflow/engine/internal/domain"
)

func (r *Runner) callVerifier(
	ctx context.Context,
	instance *domain.InstanceView,
	stepDef *domain.DefinitionStep,
	step *domain.StepView,
) {

}
