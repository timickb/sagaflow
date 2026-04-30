package domain

import "time"

// VerificationRequest - запрос сверки данных во внешнем источнике.
type VerificationRequest struct {
	Query          string
	InitialContext InstanceContext
	RuntimeContext InstanceContext
	Timeout        time.Duration
}
