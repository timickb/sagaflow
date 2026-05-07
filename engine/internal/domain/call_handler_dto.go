package domain

import "github.com/google/uuid"

type CallHandlerRequest struct {
	Service string
	Method  string

	SagaInstanceId uuid.UUID
	StepId         string
	Attempt        int
	IdempotencyKey string

	InputData InstanceContext
}
