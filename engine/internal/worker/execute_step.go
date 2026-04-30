package worker

import (
	"context"

	"github.com/timickb/sagaflow/engine/internal/domain"
)

func (r *Runner) executeStep(
	ctx context.Context,
	instance *domain.InstanceView,
	stepDef *domain.DefinitionStep,
	step *domain.StepView,
) {
	switch stepDef.Kind {
	case domain.StepKindAction, domain.StepKindCompensate, domain.StepKindReconcile:
		r.callHandler(ctx, instance, stepDef, step)
	case domain.StepKindVerify:
		r.callVerifier(ctx, instance, stepDef, step)
	}
}
