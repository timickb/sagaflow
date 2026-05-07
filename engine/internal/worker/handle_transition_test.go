package worker

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/lib/broker"
	"github.com/timickb/sagaflow/lib/utils"
)

var testRunner = &Runner{}

func TestHandleRejectedTransition(t *testing.T) {
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("action step rejected - creates next step", func(t *testing.T) {
		sagaID := uuid.New()
		initialCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{"order_id": "123"})
		require.NoError(t, err)
		runtimeCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{"user_id": "456"})
		require.NoError(t, err)

		event := &broker.SagaStepResultEvent{
			Ref: broker.SagaStepRef{
				SagaId:   sagaID,
				StepName: "validate_step",
			},
			Result: map[string]any{"validation_status": "rejected"},
		}

		sagaDef := &domain.SagaDefinition{
			Name:    "Test Saga",
			Version: 1,
			Steps: []*domain.DefinitionStep{
				{
					Id:      "validate_step",
					Kind:    domain.StepKindAction,
					Timeout: 30 * time.Second,
					Transitions: map[domain.StepOutcome]string{
						domain.OutcomeRejected:  "fail_step",
						domain.OutcomeCommitted: "process_step",
					},
				},
				{
					Id:      "fail_step",
					Kind:    domain.StepKindTerminal,
					Result:  utils.Ptr(domain.SagaResultFailed),
					Timeout: 0,
				},
				{
					Id:      "process_step",
					Kind:    domain.StepKindAction,
					Timeout: 30 * time.Second,
					Transitions: map[domain.StepOutcome]string{
						domain.OutcomeCommitted: "terminal",
					},
				},
				{
					Id:     "terminal",
					Kind:   domain.StepKindTerminal,
					Result: utils.Ptr(domain.SagaResultCompleted),
				},
			},
		}
		currentStepDef := sagaDef.Steps[0]
		instance := &domain.InstanceView{
			SagaId:         sagaID,
			Status:         domain.InstanceStatusRunning,
			InitialContext: initialCtx,
			RuntimeContext: runtimeCtx,
		}
		currentStep := &domain.StepView{
			Name:      "validate_step",
			Order:     1,
			UpdatedAt: baseTime,
		}

		result, err := testRunner.handleRejectedTransition(event, sagaDef, currentStepDef, instance, currentStep)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, result.InstanceTransitionDto)
		require.Equal(t, sagaID, result.InstanceTransitionDto.Id)
		require.Equal(t, "fail_step", result.InstanceTransitionDto.NextStepName)
		require.NotNil(t, result.StepCreateDto)
		require.Equal(t, "fail_step", result.StepCreateDto.StepName)
		require.Equal(t, 2, result.StepCreateDto.StepOrder)
		require.NotNil(t, result.StepUpdateDto)
		require.Equal(t, domain.StepStatusFailed, *result.StepUpdateDto.Status)
	})

	t.Run("verify step rejected - transitions to committed path", func(t *testing.T) {
		sagaID := uuid.New()
		runtimeCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)
		initialCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)

		event := &broker.SagaStepResultEvent{
			Ref: broker.SagaStepRef{
				SagaId:   sagaID,
				StepName: "verify_step",
			},
		}

		sagaDef := &domain.SagaDefinition{
			Steps: []*domain.DefinitionStep{
				{
					Id:      "verify_step",
					Kind:    domain.StepKindVerify,
					Timeout: 10 * time.Second,
					Transitions: map[domain.StepOutcome]string{
						domain.OutcomeCommitted: "success_step",
						domain.OutcomeUnmatched: "fail_step",
					},
				},
				{
					Id:     "success_step",
					Kind:   domain.StepKindTerminal,
					Result: utils.Ptr(domain.SagaResultCompleted),
				},
				{
					Id:     "fail_step",
					Kind:   domain.StepKindTerminal,
					Result: utils.Ptr(domain.SagaResultFailed),
				},
			},
		}
		currentStepDef := sagaDef.Steps[0]
		instance := &domain.InstanceView{
			SagaId:         sagaID,
			Status:         domain.InstanceStatusVerifying,
			InitialContext: initialCtx,
			RuntimeContext: runtimeCtx,
		}
		currentStep := &domain.StepView{
			Name:      "verify_step",
			Order:     1,
			UpdatedAt: baseTime,
		}

		result, err := testRunner.handleRejectedTransition(event, sagaDef, currentStepDef, instance, currentStep)
		require.NoError(t, err)
		require.NotNil(t, result)
		// For verify step, rejected uses OutcomeCommitted as needed outcome
		require.Equal(t, "success_step", result.InstanceTransitionDto.NextStepName)
	})

	t.Run("no transition for rejected outcome - returns no terminal state", func(t *testing.T) {
		sagaID := uuid.New()
		runtimeCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)
		initialCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)

		event := &broker.SagaStepResultEvent{
			Ref: broker.SagaStepRef{
				SagaId:   sagaID,
				StepName: "step1",
			},
		}

		sagaDef := &domain.SagaDefinition{
			Steps: []*domain.DefinitionStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindAction,
					Timeout: 30 * time.Second,
					Transitions: map[domain.StepOutcome]string{
						domain.OutcomeCommitted: "step2",
					},
				},
			},
		}
		currentStepDef := sagaDef.Steps[0]
		instance := &domain.InstanceView{
			SagaId:         sagaID,
			Status:         domain.InstanceStatusRunning,
			InitialContext: initialCtx,
			RuntimeContext: runtimeCtx,
		}
		currentStep := &domain.StepView{
			Name:      "step1",
			Order:     1,
			UpdatedAt: baseTime,
		}

		result, err := testRunner.handleRejectedTransition(event, sagaDef, currentStepDef, instance, currentStep)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Nil(t, result.StepCreateDto)
		require.Equal(t, domain.InstanceStatusInconsistent, *result.InstanceTransitionDto.Status)
	})

	t.Run("next step not found - returns failed result", func(t *testing.T) {
		sagaID := uuid.New()
		runtimeCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)
		initialCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)

		event := &broker.SagaStepResultEvent{
			Ref: broker.SagaStepRef{
				SagaId:   sagaID,
				StepName: "step1",
			},
		}

		sagaDef := &domain.SagaDefinition{
			Steps: []*domain.DefinitionStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindAction,
					Timeout: 30 * time.Second,
					Transitions: map[domain.StepOutcome]string{
						domain.OutcomeRejected: "nonexistent_step",
					},
				},
			},
		}
		currentStepDef := sagaDef.Steps[0]
		instance := &domain.InstanceView{
			SagaId:         sagaID,
			Status:         domain.InstanceStatusRunning,
			InitialContext: initialCtx,
			RuntimeContext: runtimeCtx,
		}
		currentStep := &domain.StepView{
			Name:      "step1",
			Order:     1,
			UpdatedAt: baseTime,
		}

		result, err := testRunner.handleRejectedTransition(event, sagaDef, currentStepDef, instance, currentStep)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Nil(t, result.StepCreateDto)
		require.Equal(t, domain.InstanceStatusFailed, *result.InstanceTransitionDto.Status)
	})

	t.Run("with error data - includes error info in result", func(t *testing.T) {
		sagaID := uuid.New()
		runtimeCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)
		initialCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)

		event := &broker.SagaStepResultEvent{
			Ref: broker.SagaStepRef{
				SagaId:   sagaID,
				StepName: "step1",
			},
			Error: &broker.ErrorInfo{
				Code:      "VALIDATION_ERROR",
				Message:   "Validation failed",
				Retriable: false,
				Details: map[string]any{
					"field": "email",
				},
			},
		}

		sagaDef := &domain.SagaDefinition{
			Steps: []*domain.DefinitionStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindAction,
					Timeout: 30 * time.Second,
					Transitions: map[domain.StepOutcome]string{
						domain.OutcomeRejected: "fail_step",
					},
				},
				{
					Id:     "fail_step",
					Kind:   domain.StepKindTerminal,
					Result: utils.Ptr(domain.SagaResultFailed),
				},
			},
		}
		currentStepDef := sagaDef.Steps[0]
		instance := &domain.InstanceView{
			SagaId:         sagaID,
			Status:         domain.InstanceStatusRunning,
			InitialContext: initialCtx,
			RuntimeContext: runtimeCtx,
		}
		currentStep := &domain.StepView{
			Name:      "step1",
			Order:     1,
			UpdatedAt: baseTime,
		}

		result, err := testRunner.handleRejectedTransition(event, sagaDef, currentStepDef, instance, currentStep)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, result.InstanceTransitionDto.ErrCode)
		require.Equal(t, string(domain.InstanceErrorCodeHandler), *result.InstanceTransitionDto.ErrCode)
	})

	t.Run("terminal next step - sets committed status", func(t *testing.T) {
		sagaID := uuid.New()
		runtimeCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)
		initialCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)

		event := &broker.SagaStepResultEvent{
			Ref: broker.SagaStepRef{
				SagaId:   sagaID,
				StepName: "step1",
			},
		}

		sagaDef := &domain.SagaDefinition{
			Steps: []*domain.DefinitionStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindAction,
					Timeout: 30 * time.Second,
					Transitions: map[domain.StepOutcome]string{
						domain.OutcomeRejected: "terminal",
					},
				},
				{
					Id:     "terminal",
					Kind:   domain.StepKindTerminal,
					Result: utils.Ptr(domain.SagaResultFailed),
				},
			},
		}
		currentStepDef := sagaDef.Steps[0]
		instance := &domain.InstanceView{
			SagaId:         sagaID,
			Status:         domain.InstanceStatusRunning,
			InitialContext: initialCtx,
			RuntimeContext: runtimeCtx,
		}
		currentStep := &domain.StepView{
			Name:      "step1",
			Order:     1,
			UpdatedAt: baseTime,
		}

		result, err := testRunner.handleRejectedTransition(event, sagaDef, currentStepDef, instance, currentStep)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, result.StepCreateDto)
		require.NotNil(t, result.StepCreateDto.Status)
		require.Equal(t, domain.StepStatusCommitted, *result.StepCreateDto.Status)
	})

	t.Run("next step with delay - sets next execution time", func(t *testing.T) {
		sagaID := uuid.New()
		runtimeCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)
		initialCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)

		delay := 5 * time.Second
		event := &broker.SagaStepResultEvent{
			Ref: broker.SagaStepRef{
				SagaId:   sagaID,
				StepName: "step1",
			},
		}

		sagaDef := &domain.SagaDefinition{
			Steps: []*domain.DefinitionStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindAction,
					Timeout: 30 * time.Second,
					Transitions: map[domain.StepOutcome]string{
						domain.OutcomeRejected: "step2",
					},
				},
				{
					Id:      "step2",
					Kind:    domain.StepKindAction,
					Timeout: 30 * time.Second,
					Delay:   &delay,
					Transitions: map[domain.StepOutcome]string{
						domain.OutcomeCommitted: "terminal",
					},
				},
				{
					Id:     "terminal",
					Kind:   domain.StepKindTerminal,
					Result: utils.Ptr(domain.SagaResultCompleted),
				},
			},
		}
		currentStepDef := sagaDef.Steps[0]
		instance := &domain.InstanceView{
			SagaId:         sagaID,
			Status:         domain.InstanceStatusRunning,
			InitialContext: initialCtx,
			RuntimeContext: runtimeCtx,
		}
		currentStep := &domain.StepView{
			Name:      "step1",
			Order:     1,
			UpdatedAt: baseTime,
		}

		result, err := testRunner.handleRejectedTransition(event, sagaDef, currentStepDef, instance, currentStep)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, result.InstanceTransitionDto.NextExecutionAt)
		// Next execution time should be approximately now + delay (within 2 seconds)
		approxExpected := time.Now().Add(delay)
		actual := *result.InstanceTransitionDto.NextExecutionAt
		diff := actual.Sub(approxExpected)
		require.True(t, diff < 2*time.Second && diff > -2*time.Second,
			"next execution time should be approximately now + delay")
	})
}

func TestHandleCommittedTransition(t *testing.T) {
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("action step committed - creates next step", func(t *testing.T) {
		sagaID := uuid.New()
		runtimeCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)
		initialCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)

		event := &broker.SagaStepResultEvent{
			Ref: broker.SagaStepRef{
				SagaId:   sagaID,
				StepName: "action_step",
			},
			Status: broker.SagaStepStatusCommitted,
			Result: map[string]any{"output": "data"},
		}

		sagaDef := &domain.SagaDefinition{
			Steps: []*domain.DefinitionStep{
				{
					Id:      "action_step",
					Kind:    domain.StepKindAction,
					Timeout: 30 * time.Second,
					Transitions: map[domain.StepOutcome]string{
						domain.OutcomeCommitted: "next_step",
					},
				},
				{
					Id:      "next_step",
					Kind:    domain.StepKindAction,
					Timeout: 30 * time.Second,
					Transitions: map[domain.StepOutcome]string{
						domain.OutcomeCommitted: "terminal",
					},
				},
				{
					Id:     "terminal",
					Kind:   domain.StepKindTerminal,
					Result: utils.Ptr(domain.SagaResultCompleted),
				},
			},
		}
		currentStepDef := sagaDef.Steps[0]
		instance := &domain.InstanceView{
			SagaId:         sagaID,
			Status:         domain.InstanceStatusRunning,
			InitialContext: initialCtx,
			RuntimeContext: runtimeCtx,
		}
		currentStep := &domain.StepView{
			Name:      "action_step",
			Order:     1,
			UpdatedAt: baseTime,
		}

		result, err := testRunner.handleCommittedTransition(event, sagaDef, currentStepDef, instance, currentStep)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, result.InstanceTransitionDto)
		require.Equal(t, sagaID, result.InstanceTransitionDto.Id)
		require.Equal(t, "next_step", result.InstanceTransitionDto.NextStepName)
		require.NotNil(t, result.StepCreateDto)
		require.Equal(t, "next_step", result.StepCreateDto.StepName)
		require.Equal(t, 2, result.StepCreateDto.StepOrder)
		require.NotNil(t, result.StepUpdateDto)
		require.Equal(t, domain.StepStatusCommitted, *result.StepUpdateDto.Status)
	})

	t.Run("verify step matched - transitions correctly", func(t *testing.T) {
		sagaID := uuid.New()
		runtimeCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)
		initialCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)

		event := &broker.SagaStepResultEvent{
			Ref: broker.SagaStepRef{
				SagaId:   sagaID,
				StepName: "verify_step",
			},
			Status: broker.SagaStepStatusCommitted,
		}

		sagaDef := &domain.SagaDefinition{
			Steps: []*domain.DefinitionStep{
				{
					Id:      "verify_step",
					Kind:    domain.StepKindVerify,
					Timeout: 10 * time.Second,
					Transitions: map[domain.StepOutcome]string{
						domain.OutcomeMatched:   "success",
						domain.OutcomeUnmatched: "fail",
					},
				},
				{
					Id:     "success",
					Kind:   domain.StepKindTerminal,
					Result: utils.Ptr(domain.SagaResultCompleted),
				},
				{
					Id:     "fail",
					Kind:   domain.StepKindTerminal,
					Result: utils.Ptr(domain.SagaResultFailed),
				},
			},
		}
		currentStepDef := sagaDef.Steps[0]
		instance := &domain.InstanceView{
			SagaId:         sagaID,
			Status:         domain.InstanceStatusVerifying,
			InitialContext: initialCtx,
			RuntimeContext: runtimeCtx,
		}
		currentStep := &domain.StepView{
			Name:      "verify_step",
			Order:     1,
			UpdatedAt: baseTime,
		}

		result, err := testRunner.handleCommittedTransition(event, sagaDef, currentStepDef, instance, currentStep)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "success", result.InstanceTransitionDto.NextStepName)
	})

	t.Run("no transition for committed outcome - returns no terminal state", func(t *testing.T) {
		sagaID := uuid.New()
		runtimeCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)
		initialCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)

		event := &broker.SagaStepResultEvent{
			Ref: broker.SagaStepRef{
				SagaId:   sagaID,
				StepName: "step1",
			},
			Status: broker.SagaStepStatusCommitted,
		}

		sagaDef := &domain.SagaDefinition{
			Steps: []*domain.DefinitionStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindAction,
					Timeout: 30 * time.Second,
					Transitions: map[domain.StepOutcome]string{
						domain.OutcomeFailed: "fail_step",
					},
				},
			},
		}
		currentStepDef := sagaDef.Steps[0]
		instance := &domain.InstanceView{
			SagaId:         sagaID,
			Status:         domain.InstanceStatusRunning,
			InitialContext: initialCtx,
			RuntimeContext: runtimeCtx,
		}
		currentStep := &domain.StepView{
			Name:      "step1",
			Order:     1,
			UpdatedAt: baseTime,
		}

		result, err := testRunner.handleCommittedTransition(event, sagaDef, currentStepDef, instance, currentStep)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Nil(t, result.StepCreateDto)
		require.Equal(t, domain.InstanceStatusInconsistent, *result.InstanceTransitionDto.Status)
	})

	t.Run("with result data - merges to runtime context", func(t *testing.T) {
		sagaID := uuid.New()
		runtimeCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)
		initialCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)

		event := &broker.SagaStepResultEvent{
			Ref: broker.SagaStepRef{
				SagaId:   sagaID,
				StepName: "step1",
			},
			Status: broker.SagaStepStatusCommitted,
			Result: map[string]any{
				"computed_value": 42.0,
				"status":         "completed",
			},
		}

		sagaDef := &domain.SagaDefinition{
			Steps: []*domain.DefinitionStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindAction,
					Timeout: 30 * time.Second,
					Transitions: map[domain.StepOutcome]string{
						domain.OutcomeCommitted: "terminal",
					},
				},
				{
					Id:     "terminal",
					Kind:   domain.StepKindTerminal,
					Result: utils.Ptr(domain.SagaResultCompleted),
				},
			},
		}
		currentStepDef := sagaDef.Steps[0]
		instance := &domain.InstanceView{
			SagaId:         sagaID,
			Status:         domain.InstanceStatusRunning,
			InitialContext: initialCtx,
			RuntimeContext: runtimeCtx,
		}
		currentStep := &domain.StepView{
			Name:      "step1",
			Order:     1,
			UpdatedAt: baseTime,
		}

		result, err := testRunner.handleCommittedTransition(event, sagaDef, currentStepDef, instance, currentStep)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, result.InstanceTransitionDto.RuntimeContext)
		val, err := result.InstanceTransitionDto.RuntimeContext.Find("computed_value")
		require.NoError(t, err)
		require.Equal(t, 42.0, val)
	})

	t.Run("next step with delay - sets next execution time", func(t *testing.T) {
		sagaID := uuid.New()
		runtimeCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)
		initialCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)

		delay := 10 * time.Second
		event := &broker.SagaStepResultEvent{
			Ref: broker.SagaStepRef{
				SagaId:   sagaID,
				StepName: "step1",
			},
			Status: broker.SagaStepStatusCommitted,
		}

		sagaDef := &domain.SagaDefinition{
			Steps: []*domain.DefinitionStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindAction,
					Timeout: 30 * time.Second,
					Transitions: map[domain.StepOutcome]string{
						domain.OutcomeCommitted: "step2",
					},
				},
				{
					Id:      "step2",
					Kind:    domain.StepKindAction,
					Timeout: 30 * time.Second,
					Delay:   &delay,
					Transitions: map[domain.StepOutcome]string{
						domain.OutcomeCommitted: "terminal",
					},
				},
				{
					Id:     "terminal",
					Kind:   domain.StepKindTerminal,
					Result: utils.Ptr(domain.SagaResultCompleted),
				},
			},
		}
		currentStepDef := sagaDef.Steps[0]
		instance := &domain.InstanceView{
			SagaId:         sagaID,
			Status:         domain.InstanceStatusRunning,
			InitialContext: initialCtx,
			RuntimeContext: runtimeCtx,
		}
		currentStep := &domain.StepView{
			Name:      "step1",
			Order:     1,
			UpdatedAt: baseTime,
		}

		result, err := testRunner.handleCommittedTransition(event, sagaDef, currentStepDef, instance, currentStep)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, result.InstanceTransitionDto.NextExecutionAt)
		// Next execution time should be approximately now + delay (within 2 seconds)
		approxExpected := time.Now().Add(delay)
		actual := *result.InstanceTransitionDto.NextExecutionAt
		diff := actual.Sub(approxExpected)
		require.True(t, diff < 2*time.Second && diff > -2*time.Second,
			"next execution time should be approximately now + delay")
	})
}

func TestHandleFailedTransition(t *testing.T) {
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("step with retry available - schedules retry", func(t *testing.T) {
		sagaID := uuid.New()
		runtimeCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)
		initialCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)

		event := &broker.SagaStepResultEvent{
			Ref: broker.SagaStepRef{
				SagaId:   sagaID,
				StepName: "retryable_step",
			},
			Error: &broker.ErrorInfo{
				Code:      "TEMPORARY_ERROR",
				Message:   "Temporary failure",
				Retriable: true,
			},
		}

		sagaDef := &domain.SagaDefinition{
			Steps: []*domain.DefinitionStep{
				{
					Id:      "retryable_step",
					Kind:    domain.StepKindAction,
					Timeout: 30 * time.Second,
					Retry: &domain.RetryPolicy{
						MaxAttempts: 3,
						Backoff:     domain.RetryBackoffTypeFixed,
						Delay:       1000 * time.Millisecond,
					},
					Transitions: map[domain.StepOutcome]string{
						domain.OutcomeCommitted: "next",
						domain.OutcomeFailed:    "fail",
					},
				},
				{
					Id:     "next",
					Kind:   domain.StepKindTerminal,
					Result: utils.Ptr(domain.SagaResultCompleted),
				},
				{
					Id:     "fail",
					Kind:   domain.StepKindTerminal,
					Result: utils.Ptr(domain.SagaResultFailed),
				},
			},
		}
		currentStepDef := sagaDef.Steps[0]
		instance := &domain.InstanceView{
			SagaId:         sagaID,
			Status:         domain.InstanceStatusRunning,
			InitialContext: initialCtx,
			RuntimeContext: runtimeCtx,
		}
		currentStep := &domain.StepView{
			Name:      "retryable_step",
			Order:     1,
			Attempt:   1,
			UpdatedAt: baseTime,
		}

		result, err := testRunner.handleFailedTransition(event, sagaDef, currentStepDef, instance, currentStep)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Nil(t, result.StepCreateDto)
		require.Equal(t, "retryable_step", result.InstanceTransitionDto.NextStepName)
		require.NotNil(t, result.InstanceTransitionDto.NextExecutionAt)
		require.Equal(t, baseTime.Add(1000*time.Millisecond), *result.InstanceTransitionDto.NextExecutionAt)
		require.NotNil(t, result.StepUpdateDto)
		require.True(t, result.StepUpdateDto.IncrementAttempt)
	})

	t.Run("exponential backoff retry", func(t *testing.T) {
		sagaID := uuid.New()
		runtimeCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)
		initialCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)

		event := &broker.SagaStepResultEvent{
			Ref: broker.SagaStepRef{
				SagaId:   sagaID,
				StepName: "retryable_step",
			},
			Error: &broker.ErrorInfo{
				Code:      "ERROR",
				Message:   "Failed",
				Retriable: true,
			},
		}

		sagaDef := &domain.SagaDefinition{
			Steps: []*domain.DefinitionStep{
				{
					Id:      "retryable_step",
					Kind:    domain.StepKindAction,
					Timeout: 30 * time.Second,
					Retry: &domain.RetryPolicy{
						MaxAttempts: 5,
						Backoff:     domain.RetryBackoffTypeExponential,
						Delay:       1000 * time.Millisecond,
					},
					Transitions: map[domain.StepOutcome]string{
						domain.OutcomeFailed: "fail",
					},
				},
				{
					Id:     "fail",
					Kind:   domain.StepKindTerminal,
					Result: utils.Ptr(domain.SagaResultFailed),
				},
			},
		}
		currentStepDef := sagaDef.Steps[0]
		instance := &domain.InstanceView{
			SagaId:         sagaID,
			Status:         domain.InstanceStatusRunning,
			InitialContext: initialCtx,
			RuntimeContext: runtimeCtx,
		}
		currentStep := &domain.StepView{
			Name:      "retryable_step",
			Order:     1,
			Attempt:   2,
			UpdatedAt: baseTime,
		}

		result, err := testRunner.handleFailedTransition(event, sagaDef, currentStepDef, instance, currentStep)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, result.InstanceTransitionDto.NextExecutionAt)
		// Exponential: delay * 2 * attempt = 1000 * 2 * 2 = 4000ms
		require.Equal(t, baseTime.Add(4000*time.Millisecond), *result.InstanceTransitionDto.NextExecutionAt)
	})

	t.Run("non-retriable error - skips to failure transition", func(t *testing.T) {
		sagaID := uuid.New()
		runtimeCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)
		initialCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)

		event := &broker.SagaStepResultEvent{
			Ref: broker.SagaStepRef{
				SagaId:   sagaID,
				StepName: "step1",
			},
			Error: &broker.ErrorInfo{
				Code:      "NON_RETRIABLE",
				Message:   "Cannot retry",
				Retriable: false,
			},
		}

		sagaDef := &domain.SagaDefinition{
			Steps: []*domain.DefinitionStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindAction,
					Timeout: 30 * time.Second,
					Retry: &domain.RetryPolicy{
						MaxAttempts: 3,
						Backoff:     domain.RetryBackoffTypeFixed,
						Delay:       1000 * time.Millisecond,
					},
					Transitions: map[domain.StepOutcome]string{
						domain.OutcomeFailed: "fail_step",
					},
				},
				{
					Id:     "fail_step",
					Kind:   domain.StepKindTerminal,
					Result: utils.Ptr(domain.SagaResultFailed),
				},
			},
		}
		currentStepDef := sagaDef.Steps[0]
		instance := &domain.InstanceView{
			SagaId:         sagaID,
			Status:         domain.InstanceStatusRunning,
			InitialContext: initialCtx,
			RuntimeContext: runtimeCtx,
		}
		currentStep := &domain.StepView{
			Name:      "step1",
			Order:     1,
			Attempt:   1,
			UpdatedAt: baseTime,
		}

		result, err := testRunner.handleFailedTransition(event, sagaDef, currentStepDef, instance, currentStep)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, result.StepCreateDto)
		require.Equal(t, "fail_step", result.StepCreateDto.StepName)
	})

	t.Run("retry exhausted - transitions to failure step", func(t *testing.T) {
		sagaID := uuid.New()
		runtimeCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)
		initialCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)

		event := &broker.SagaStepResultEvent{
			Ref: broker.SagaStepRef{
				SagaId:   sagaID,
				StepName: "step1",
			},
			Error: &broker.ErrorInfo{
				Code:      "ERROR",
				Message:   "Failed",
				Retriable: true,
			},
		}

		sagaDef := &domain.SagaDefinition{
			Steps: []*domain.DefinitionStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindAction,
					Timeout: 30 * time.Second,
					Retry: &domain.RetryPolicy{
						MaxAttempts: 3,
						Backoff:     domain.RetryBackoffTypeFixed,
						Delay:       1000 * time.Millisecond,
					},
					Transitions: map[domain.StepOutcome]string{
						domain.OutcomeFailed: "fail_step",
					},
				},
				{
					Id:     "fail_step",
					Kind:   domain.StepKindTerminal,
					Result: utils.Ptr(domain.SagaResultFailed),
				},
			},
		}
		currentStepDef := sagaDef.Steps[0]
		instance := &domain.InstanceView{
			SagaId:         sagaID,
			Status:         domain.InstanceStatusRunning,
			InitialContext: initialCtx,
			RuntimeContext: runtimeCtx,
		}
		currentStep := &domain.StepView{
			Name:      "step1",
			Order:     1,
			Attempt:   4, // Exhausted: attempt > max_attempts
			UpdatedAt: baseTime,
		}

		result, err := testRunner.handleFailedTransition(event, sagaDef, currentStepDef, instance, currentStep)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, result.StepCreateDto)
		require.Equal(t, "fail_step", result.StepCreateDto.StepName)
		require.NotNil(t, result.StepUpdateDto)
		require.Equal(t, domain.StepStatusFailed, *result.StepUpdateDto.Status)
	})

	t.Run("no failure transition - returns no terminal state", func(t *testing.T) {
		sagaID := uuid.New()
		runtimeCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)
		initialCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)

		event := &broker.SagaStepResultEvent{
			Ref: broker.SagaStepRef{
				SagaId:   sagaID,
				StepName: "step1",
			},
			Error: &broker.ErrorInfo{
				Code:      "ERROR",
				Message:   "Failed",
				Retriable: true,
			},
		}

		sagaDef := &domain.SagaDefinition{
			Steps: []*domain.DefinitionStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindAction,
					Timeout: 30 * time.Second,
					Retry: &domain.RetryPolicy{
						MaxAttempts: 1,
						Backoff:     domain.RetryBackoffTypeFixed,
						Delay:       1000 * time.Millisecond,
					},
					Transitions: map[domain.StepOutcome]string{
						domain.OutcomeCommitted: "next",
					},
				},
			},
		}
		currentStepDef := sagaDef.Steps[0]
		instance := &domain.InstanceView{
			SagaId:         sagaID,
			Status:         domain.InstanceStatusRunning,
			InitialContext: initialCtx,
			RuntimeContext: runtimeCtx,
		}
		currentStep := &domain.StepView{
			Name:      "step1",
			Order:     1,
			Attempt:   2, // Exhausted: attempt >= max_attempts
			UpdatedAt: baseTime,
		}

		result, err := testRunner.handleFailedTransition(event, sagaDef, currentStepDef, instance, currentStep)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Nil(t, result.StepCreateDto)
		require.Equal(t, domain.InstanceStatusInconsistent, *result.InstanceTransitionDto.Status)
	})

	t.Run("terminal next step - sets committed status", func(t *testing.T) {
		sagaID := uuid.New()
		runtimeCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)
		initialCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)

		event := &broker.SagaStepResultEvent{
			Ref: broker.SagaStepRef{
				SagaId:   sagaID,
				StepName: "step1",
			},
			Error: &broker.ErrorInfo{
				Code:      "ERROR",
				Message:   "Failed",
				Retriable: true,
			},
		}

		sagaDef := &domain.SagaDefinition{
			Steps: []*domain.DefinitionStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindAction,
					Timeout: 30 * time.Second,
					// No retry policy - goes directly to failure transition
					Transitions: map[domain.StepOutcome]string{
						domain.OutcomeFailed: "terminal",
					},
				},
				{
					Id:     "terminal",
					Kind:   domain.StepKindTerminal,
					Result: utils.Ptr(domain.SagaResultFailed),
				},
			},
		}
		currentStepDef := sagaDef.Steps[0]
		instance := &domain.InstanceView{
			SagaId:         sagaID,
			Status:         domain.InstanceStatusRunning,
			InitialContext: initialCtx,
			RuntimeContext: runtimeCtx,
		}
		currentStep := &domain.StepView{
			Name:      "step1",
			Order:     1,
			Attempt:   1,
			UpdatedAt: baseTime,
		}

		result, err := testRunner.handleFailedTransition(event, sagaDef, currentStepDef, instance, currentStep)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, result.StepCreateDto)
		require.NotNil(t, result.StepCreateDto.Status)
		require.Equal(t, domain.StepStatusCommitted, *result.StepCreateDto.Status)
	})
}
