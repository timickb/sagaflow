package domain

// DataSourceType - тип подключения к источнику данных для сверочных операций.
type DataSourceType string

const (
	DataSourceTypePostgres   DataSourceType = "postgres"
	DataSourceTypeMySQL      DataSourceType = "mysql"
	DataSourceTypeMongo      DataSourceType = "mongo"
	DataSourceTypeClickHouse DataSourceType = "clickhouse"
	DataSourceTypeRest       DataSourceType = "rest"
)
