package verification_source

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/timickb/sagaflow/engine/internal/domain"
)

var sqlParamRegexp = regexp.MustCompile(`\$\{(input|runtime)\.([^}]+)}`)

type SQLDialect string

const (
	SQLDialectPostgres   SQLDialect = "postgres"
	SQLDialectClickHouse SQLDialect = "clickhouse"
	SQLDialectMySQL      SQLDialect = "mysql"
)

type SQLSource struct {
	db      *sql.DB
	dialect SQLDialect
}

func NewSQLSource(db *sql.DB, dialect SQLDialect) *SQLSource {
	return &SQLSource{
		db:      db,
		dialect: dialect,
	}
}

func (s *SQLSource) Verify(ctx context.Context, req *domain.VerificationRequest) (*domain.VerificationResult, error) {
	query, args, err := s.buildQuery(req)
	if err != nil {
		return nil, err
	}

	row := s.db.QueryRowContext(ctx, query, args...)

	var actual any
	if err = row.Scan(&actual); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &domain.VerificationResult{
				Verified: false,
			}, nil
		}

		return nil, fmt.Errorf("execute verification query: %w", err)
	}

	return &domain.VerificationResult{
		Verified:     compareScalar(actual, req.Expected),
		ActualResult: actual,
	}, nil
}

func (s *SQLSource) buildQuery(req *domain.VerificationRequest) (string, []any, error) {
	var args []any
	var buildErr error

	query := sqlParamRegexp.ReplaceAllStringFunc(req.Query, func(match string) string {
		if buildErr != nil {
			return s.placeholder(len(args) + 1)
		}

		parts := sqlParamRegexp.FindStringSubmatch(match)
		if len(parts) != 3 {
			buildErr = fmt.Errorf("invalid verification query parameter %q", match)
			return s.placeholder(len(args) + 1)
		}

		source := parts[1]
		path := parts[2]

		value, err := findVerificationParam(req, source, path)
		if err != nil {
			buildErr = err
			return s.placeholder(len(args) + 1)
		}

		args = append(args, value)

		return s.placeholder(len(args))
	})
	if buildErr != nil {
		return "", nil, buildErr
	}

	if strings.Contains(query, "${") {
		return "", nil, fmt.Errorf("verification query contains invalid parameter expression")
	}

	return query, args, nil
}

func (s *SQLSource) placeholder(argNumber int) string {
	switch s.dialect {
	case SQLDialectPostgres:
		return fmt.Sprintf("$%d", argNumber)
	case SQLDialectClickHouse, SQLDialectMySQL:
		return "?"
	default:
		return "?"
	}
}

func compareScalar(actual any, expected any) bool {
	return fmt.Sprint(actual) == fmt.Sprint(expected)
}
