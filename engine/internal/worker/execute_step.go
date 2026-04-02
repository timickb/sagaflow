package worker

import (
	"context"

	"github.com/timickb/sagaflow/engine/internal/domain"
)

func (r *Runner) executeStep(ctx context.Context, instance *domain.InstanceView, stepDef *domain.DefinitionStep) {
	switch stepDef.Kind {
	case domain.StepKindAction, domain.StepKindCompensate:
		r.callHandler(ctx, instance, stepDef)
	}
}
