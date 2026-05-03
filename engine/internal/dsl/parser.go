package dsl

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/lib/utils"
)

var (
	dstVarRegexp    = regexp.MustCompile(`^[a-zA-Z_-][a-zA-Z0-9_-]*$`)
	inputSrcRegexp  = regexp.MustCompile(`^\$\{(?:runtime|input)\.[a-zA-Z_-]+(?:\.[a-zA-Z_-]+)*\}$`)
	outputSrcRegexp = regexp.MustCompile(`^\$\{(?:result|error)\.[a-zA-Z_-]+(?:\.[a-zA-Z_-]+)*\}$`)
)

// ValidateAndNormalize - валидировать модель саги и преобразовать в доменную модель
func (r *RawSagaDefinition) ValidateAndNormalize() (*domain.SagaDefinition, error) {
	if r.Saga.Version <= 0 {
		return nil, errors.New("definitions version must be positive number")
	}
	if r.Saga.Name == "" {
		return nil, errors.New("saga name could not be empty")
	}

	allStepIds := utils.MapSlice(r.Steps, func(step RawStep) string {
		return step.Id
	})
	if !utils.Contains(allStepIds, r.Saga.Start) {
		return nil, fmt.Errorf("start step %s does not declared", r.Saga.Start)
	}

	steps := make([]*domain.DefinitionStep, 0, len(r.Steps))
	for _, rawStep := range r.Steps {
		step, err := parseStep(rawStep)
		if err != nil {
			return nil, fmt.Errorf("parse step: %w", err)
		}
		steps = append(steps, step)
	}

	for _, step := range steps {
		if step.CompensateStepId != "" {
			if !utils.Contains(allStepIds, step.CompensateStepId) {
				return nil, fmt.Errorf("compensate step %s does not declared", step.Id)
			}
		}
		for _, nextStep := range step.Transitions {
			if !utils.Contains(allStepIds, nextStep) {
				return nil, fmt.Errorf("next step %s does not declared", step.Id)
			}
		}
	}

	stepById := utils.SliceToMap(steps, func(step *domain.DefinitionStep) (string, *domain.DefinitionStep) {
		return step.Id, step
	})

	return &domain.SagaDefinition{
		Name:        r.Saga.Name,
		Version:     r.Saga.Version,
		StartStepId: r.Saga.Start,
		Steps:       steps,
		StepById:    stepById,
	}, nil
}

func parseStep(step RawStep) (*domain.DefinitionStep, error) {
	parsedStep := &domain.DefinitionStep{
		Id:   step.Id,
		Kind: step.Kind,
	}
	if step.Id == "" {
		return nil, errors.New("step id could not be empty")
	}

	// Таймаут обязателен для любого шага кроме терминального
	if step.Kind != domain.StepKindTerminal {
		timeout, err := time.ParseDuration(step.Timeout)
		if err != nil {
			return nil, fmt.Errorf("parse step timeout: %w", err)
		}
		if timeout == 0 {
			return nil, errors.New("step timeout is zero")
		}
		parsedStep.Timeout = timeout
	}

	// Время отложенного запуска
	if !utils.IsStrNilOrEmpty(step.Delay) {
		delay, err := time.ParseDuration(*step.Delay)
		if err != nil {
			return nil, fmt.Errorf("parse step delay: %w", err)
		}
		if delay == 0 {
			return nil, errors.New("step delay is zero")
		}
		parsedStep.Delay = &delay
	}

	// Конфигурация повторных попыток
	if step.Retry != nil {
		if !step.Retry.Backoff.IsValid() {
			return nil, errors.New("invalid step retry backoff type")
		}
		if step.Retry.MaxAttempts <= 0 {
			return nil, errors.New("step retry max attempts must be positive")
		}
		delay, pErr := time.ParseDuration(step.Retry.Delay)
		if pErr != nil {
			return nil, fmt.Errorf("parse step retry delay: %w", pErr)
		}
		parsedStep.Retry = &domain.RetryPolicy{
			MaxAttempts: step.Retry.MaxAttempts,
			Backoff:     step.Retry.Backoff,
			Delay:       delay,
		}
	}

	// Входные данные
	if step.Input != nil {
		parsedStep.Inputs = make([]domain.StepInputParam, 0, len(step.Input))
		for dst, src := range step.Input {
			if !inputSrcRegexp.MatchString(src) {
				return nil, fmt.Errorf("invalid input source %s", src)
			}
			if !dstVarRegexp.MatchString(dst) {
				return nil, fmt.Errorf("invalid input destination %s", dst)
			}
			src = src[2 : len(src)-1]
			srcParts := splitFirst(src, ".")
			parsedStep.Inputs = append(parsedStep.Inputs, domain.StepInputParam{
				SourceNamespace:  domain.StepInputSource(srcParts[0]),
				SourcePath:       srcParts[1],
				DestinationParam: dst,
			})
		}
	}

	// Выходные данные
	if step.Output != nil {
		parsedStep.Outputs = make([]domain.StepOutputParam, 0, len(step.Output))
		for dst, src := range step.Output {
			if !outputSrcRegexp.MatchString(src) {
				return nil, fmt.Errorf("invalid output source %s", src)
			}
			if !dstVarRegexp.MatchString(dst) {
				return nil, fmt.Errorf("invalid output destination %s", dst)
			}
			src = src[2 : len(src)-1]
			srcParts := splitFirst(src, ".")
			parsedStep.Outputs = append(parsedStep.Outputs, domain.StepOutputParam{
				SourceNamespace:  domain.StepOutputSource(srcParts[0]),
				SourceParam:      srcParts[1],
				DestinationParam: dst,
			})
		}
	}

	switch step.Kind {
	case domain.StepKindAction, domain.StepKindCompensate, domain.StepKindReconcile:
		if step.Handler == nil {
			return nil, errors.New("handler is required for action/reconcile step")
		}
		if step.Verifier != nil {
			return nil, errors.New("unexpected verifier for action/reconcile step")
		}
		if step.Result != nil {
			return nil, errors.New("unexpected result for action/reconcile step")
		}
		parsedStep.Handler = &domain.Handler{
			Service: step.Handler.Service,
			Method:  step.Handler.Method,
		}
	case domain.StepKindVerify:
		if step.Verifier == nil {
			return nil, errors.New("verifier is required for verify step")
		}
		if step.Handler != nil {
			return nil, errors.New("unexpected handler for verify step")
		}
		if step.Result != nil {
			return nil, errors.New("unexpected result for verify step")
		}
		if !step.Verifier.Type.IsValid() {
			return nil, errors.New("invalid verifier type")
		}
		parsedStep.Verifier = &domain.Verifier{
			Type:       step.Verifier.Type,
			Datasource: step.Verifier.Datasource,
			Query:      step.Verifier.Query,
			Expect: domain.VerifierExpectModel{
				Equals: step.Verifier.Expect.Equals,
			},
		}
	case domain.StepKindTerminal:
		if step.Result == nil || !(*step.Result).IsValid() {
			return nil, errors.New("invalid result value for terminal step")
		}
		if step.Handler != nil {
			return nil, errors.New("unexpected handler for terminal step")
		}
		if step.Verifier != nil {
			return nil, errors.New("unexpected verifier for terminal step")
		}
		parsedStep.Result = step.Result
	}

	transitions := make(map[domain.StepOutcome]string, len(step.On))
	for event, nextStep := range step.On {
		if !domain.StepOutcome(event).IsValid() {
			return nil, fmt.Errorf("invalid step outcome: %s", event)
		}
		if nextStep == "" {
			return nil, fmt.Errorf("unexpected empty next step")
		}
		transitions[domain.StepOutcome(event)] = nextStep
	}
	parsedStep.Transitions = transitions

	return parsedStep, nil
}
