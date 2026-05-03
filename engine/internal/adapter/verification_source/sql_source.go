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

type SQLSource struct {
	db *sql.DB
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
			return "?"
		}

		parts := sqlParamRegexp.FindStringSubmatch(match)
		if len(parts) != 3 {
			buildErr = fmt.Errorf("invalid verification query parameter %q", match)
			return "?"
		}

		source := parts[1]
		path := parts[2]

		value, err := findVerificationParam(req, source, path)
		if err != nil {
			buildErr = err
			return "?"
		}

		args = append(args, value)
		return "?"
	})
	if buildErr != nil {
		return "", nil, buildErr
	}

	if strings.Contains(query, "${") {
		return "", nil, fmt.Errorf("verification query contains invalid parameter expression")
	}

	return query, args, nil
}

func compareScalar(actual any, expected any) bool {
	return fmt.Sprint(actual) == fmt.Sprint(expected)
}
