package domain

import (
	"time"
)

// VerificationSourceParams - параметры источника данных
type VerificationSourceParams struct {
	Type    VerificationSourceType
	Timeout time.Duration

	DSN string
	// BaseUrl - префикс для URL запросов (для type=api)
	BaseUrl string
	// Headers - заголовки для HTTP-запросов (для type=api)
	Headers map[string]string
}
