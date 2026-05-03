package verification_source

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/timickb/sagaflow/engine/internal/domain"
)

func NewVerificationSource(params *domain.VerificationSourceParams) (domain.VerificationSource, error) {
	switch params.Type {
	case domain.VerificationSourceTypePostgres:
		db, err := sql.Open("pgx", params.DSN)
		if err != nil {
			return nil, err
		}
		return &SQLSource{db: db}, nil

	case domain.VerificationSourceTypeClickHouse:
		db, err := sql.Open("clickhouse", params.DSN)
		if err != nil {
			return nil, err
		}
		return &SQLSource{db: db}, nil

	case domain.VerificationSourceTypeRest:
		return &RESTSource{
			client:  &http.Client{Timeout: params.Timeout},
			baseUrl: params.BaseUrl,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported verification source type: %s", params.Type)
	}
}

func findVerificationParam(req *domain.VerificationRequest, source, path string) (any, error) {
	switch source {
	case "input":
		value, err := req.InitialContext.Find(path)
		if err != nil {
			return nil, fmt.Errorf("input parameter %q not found: %w", path, err)
		}

		return value, nil
	case "runtime":
		value, err := req.RuntimeContext.Find(path)
		if err != nil {
			return nil, fmt.Errorf("runtime parameter %q not found: %w", path, err)
		}

		return value, nil
	default:
		return nil, fmt.Errorf("unsupported verification parameter source %q", source)
	}
}
