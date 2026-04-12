package domain

import "context"

type Transactor interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}
