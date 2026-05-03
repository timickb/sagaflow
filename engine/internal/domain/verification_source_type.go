package domain

// VerificationSourceType - тип подключения к источнику данных для сверочных операций.
type VerificationSourceType string

const (
	VerificationSourceTypePostgres   VerificationSourceType = "postgres"
	VerificationSourceTypeClickHouse VerificationSourceType = "clickhouse"
	VerificationSourceTypeRest       VerificationSourceType = "rest"
)
