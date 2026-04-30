package worker

import (
	"time"

	"github.com/timickb/sagaflow/engine/internal/domain"
)

func calculateNextRetry(retryPolicy *domain.RetryPolicy, currentStep *domain.StepView) time.Time {
	delay := retryPolicy.Delay.Milliseconds()
	if retryPolicy.Backoff == domain.RetryBackoffTypeExponential {
		delay *= 2 * int64(currentStep.Attempt)
	}
	return currentStep.UpdatedAt.Add(time.Duration(delay) * time.Millisecond)
}
