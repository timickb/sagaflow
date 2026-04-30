package verification_source

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/timickb/sagaflow/engine/internal/domain"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func NewVerificationSource(params domain.VerificationSourceParams) (domain.VerificationSource, error) {
	switch params.Type {
	case domain.DataSourceTypePostgres:
		db, err := sql.Open("pgx", params.DSN)
		if err != nil {
			return nil, err
		}
		return &SQLSource{db: db}, nil

	case domain.DataSourceTypeClickHouse:
		db, err := sql.Open("clickhouse", params.DSN)
		if err != nil {
			return nil, err
		}
		return &SQLSource{db: db}, nil

	case domain.DataSourceTypeRest:
		return &RESTSource{
			client:  &http.Client{Timeout: params.Timeout},
			baseUrl: params.BaseUrl,
		}, nil

	case domain.DataSourceTypeMongo:
		client, err := mongo.Connect(options.Client().ApplyURI(params.Uri))
		if err != nil {
			return nil, err
		}
		return &MongoSource{client: client}, nil

	default:
		return nil, fmt.Errorf("unsupported verification source type: %s", params.Type)
	}
}
