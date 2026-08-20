package postgres

import (
	"context"
	"iter"

	"github.com/jackc/pgx/v5"
)

// Queryer is implemented by both pgxpool.Pool and pgx.Tx.
type Queryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

// ScanFunc maps a database row to a typed data-layer model.
type ScanFunc[T any] func(pgx.CollectableRow) (T, error)

// Result keeps iteration lazy while making query and scan errors explicit.
type Result[T any] struct {
	Value T
	Err   error
}

// QueryRepository is a generic read repository for repeated SQL-to-model plumbing.
// Domain repositories remain specific and express aggregate operations explicitly.
type QueryRepository[T any] struct {
	queryer Queryer
	query   string
	scan    ScanFunc[T]
}

func NewQueryRepository[T any](queryer Queryer, query string, scan ScanFunc[T]) *QueryRepository[T] {
	return &QueryRepository[T]{queryer: queryer, query: query, scan: scan}
}

// Seq materializes the current database batch, closes the cursor, and then yields
// typed values. Closing before yield allows consumers to execute updates on the
// same transaction connection while iterating.
func (r *QueryRepository[T]) Seq(ctx context.Context, args ...any) iter.Seq[Result[T]] {
	return func(yield func(Result[T]) bool) {
		rows, err := r.queryer.Query(ctx, r.query, args...)
		if err != nil {
			yield(Result[T]{Err: err})
			return
		}
		values := make([]T, 0)
		for rows.Next() {
			value, scanErr := r.scan(rows)
			if scanErr != nil {
				rows.Close()
				yield(Result[T]{Err: scanErr})
				return
			}
			values = append(values, value)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			yield(Result[T]{Err: err})
			return
		}
		rows.Close()
		for _, value := range values {
			if !yield(Result[T]{Value: value}) {
				return
			}
		}
	}
}
