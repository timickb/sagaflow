package domain

import "time"

// VerificationSourceParams - параметры источника данных
type VerificationSourceParams struct {
	Type    DataSourceType `json:"type"`
	Timeout time.Duration  `json:"timeout"`

	DSN     string `json:"dsn,omitempty"`
	BaseUrl string `json:"base_url,omitempty"`
	Uri     string `json:"uri,omitempty"`
}
