package postgres

import (
	"context"
	"iter"

	"github.com/jackc/pgx/v5"
)

// Queryer задаёт общий контракт выполнения запросов для pgxpool.Pool и pgx.Tx.
type Queryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

// ScanFunc преобразует строку БД в типизированную модель слоя данных.
type ScanFunc[T any] func(pgx.CollectableRow) (T, error)

// Result хранит очередное значение итератора или ошибку чтения.
type Result[T any] struct {
	Value T
	Err   error
}

// QueryRepository устраняет повторяющийся код выборки и преобразования строк БД.
// Доменные репозитории при этом остаются специализированными и явно выражают
// операции над агрегатами.
type QueryRepository[T any] struct {
	queryer Queryer
	query   string
	scan    ScanFunc[T]
}

// NewQueryRepository создаёт типизированный репозиторий для заданного SQL-запроса.
func NewQueryRepository[T any](queryer Queryer, query string, scan ScanFunc[T]) *QueryRepository[T] {
	return &QueryRepository[T]{queryer: queryer, query: query, scan: scan}
}

// Seq загружает текущую пачку, закрывает курсор и затем выдаёт типизированные
// значения. Закрытие курсора до передачи значений позволяет выполнять UPDATE
// в той же транзакции во время обхода последовательности.
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
