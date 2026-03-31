package dsl

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/timickb/sagaflow/engine/pkg/utils"
	"gopkg.in/yaml.v3"
)

var (
	allowedDSLExtensions = []string{"yaml", "yml"}
)

func readRawSagas(path string) ([]*RawSagaDefinition, error) {
	entries, _ := os.ReadDir(path)
	handledSagas := make([]*RawSagaDefinition, 0)

	for _, entry := range entries {
		if entry.IsDir() {
			children, err := readRawSagas(path + "/" + entry.Name())
			if err != nil {
				return nil, err
			}
			handledSagas = append(handledSagas, children...)
		} else {
			scenario, err := readRawSaga(path, entry)
			if err != nil {
				return nil, err
			}
			handledSagas = append(handledSagas, scenario)
		}
	}

	return handledSagas, nil
}

func readRawSaga(path string, entry os.DirEntry) (*RawSagaDefinition, error) {
	nameParts := strings.Split(entry.Name(), ".")
	if len(nameParts) < 3 {
		return nil, errors.New("invalid DSL file name: correct template is <saga_name>.<saga_version>.yaml|yml")
	}
	if !utils.Contains(allowedDSLExtensions, nameParts[len(nameParts)-1]) {
		return nil, fmt.Errorf("invalid DSL file extension: %s", nameParts[len(nameParts)-1])
	}

	content, err := os.ReadFile(path + "/" + entry.Name())
	if err != nil {
		return nil, err
	}

	var raw RawSagaDefinition
	if err = yaml.Unmarshal(content, &raw); err != nil {
		return nil, err
	}
	return &raw, nil
}

func (r *RawSagaDefinition) ValidateAndNormalize() (*Definition, error) {
	if r.Saga.Version <= 0 {
		return nil, errors.New("sagas version must be positive number")
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

	steps := make([]*Step, 0, len(r.Steps))
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

	stepById := utils.SliceToMap(steps, func(step *Step) (string, *Step) {
		return step.Id, step
	})

	return &Definition{
		Name:        r.Saga.Name,
		Version:     r.Saga.Version,
		StartStepId: r.Saga.Start,
		Steps:       steps,
		StepById:    stepById,
	}, nil
}

func parseStep(step RawStep) (*Step, error) {
	parsedStep := &Step{
		Id:   step.Id,
		Kind: step.Kind,
	}
	if step.Id == "" {
		return nil, errors.New("step id could not be empty")
	}

	timeout, err := time.ParseDuration(step.Timeout)
	if err != nil {
		return nil, fmt.Errorf("parse step timeout: %w", err)
	}
	if timeout == 0 {
		return nil, errors.New("step timeout is zero")
	}
	parsedStep.Timeout = timeout

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
		parsedStep.Retry = &RetryPolicy{
			MaxAttempts: step.Retry.MaxAttempts,
			Backoff:     step.Retry.Backoff,
			Delay:       delay,
		}
	}

	switch step.Kind {
	case StepKindAction, StepKindReconcile:
		if step.Handler == nil {
			return nil, errors.New("handler is required for action/reconcile step")
		}
		if step.Verifier != nil {
			return nil, errors.New("unexpected verifier for action/reconcile step")
		}
		if step.Result != nil {
			return nil, errors.New("unexpected result for action/reconcile step")
		}
		parsedStep.Handler = &Handler{
			Service: step.Handler.Service,
			Method:  step.Handler.Method,
		}
	case StepKindVerify:
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
		parsedStep.Verifier = &Verifier{
			Type:       step.Verifier.Type,
			Datasource: step.Verifier.Datasource,
			Query:      step.Verifier.Query,
			Checks:     step.Verifier.Checks,
		}
	case StepKindTerminal:
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

	transitions := make(map[StepOutcome]string, len(step.On))
	for event, nextStep := range step.On {
		if !StepOutcome(event).IsValid() {
			return nil, fmt.Errorf("invalid step outcome: %s", event)
		}
		if nextStep == "" {
			return nil, fmt.Errorf("unexpected empty next step")
		}
		transitions[StepOutcome(event)] = nextStep
	}
	parsedStep.Transitions = transitions

	return parsedStep, nil
}
