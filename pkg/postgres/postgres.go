package postgres

import (
	"context"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	defaultMaxPoolSize  = 1
	defaultConnAttempts = 10
	defaultConnTimeout  = time.Second
)

// txKey используется для хранения транзакции в контексте.
type txKey struct{}

// injectTx добавляет транзакцию в контекст.
func injectTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// extractTx извлекает транзакцию из контекста, если она там присутствует.
func extractTx(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	return tx, ok
}

// QueryRunner является общим интерфейсом для методов, выполняющих запросы.
type QueryRunner interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type PgxPool interface {
	Close()
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
	Ping(ctx context.Context) error
}

type Postgres struct {
	maxPoolSize  int
	connAttempts int
	connTimeout  time.Duration

	Builder squirrel.StatementBuilderType
	Pool    *pgxpool.Pool
}

func New(url string, opts ...Option) (*Postgres, error) {
	pg := &Postgres{
		maxPoolSize:  defaultMaxPoolSize,
		connAttempts: defaultConnAttempts,
		connTimeout:  defaultConnTimeout,
	}

	// Custom options
	for _, opt := range opts {
		opt(pg)
	}

	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	poolConfig, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("postgres - NewPostgres - pgxpool.ParseConfig: %w", err)
	}

	poolConfig.MaxConns = int32(pg.maxPoolSize)

	for pg.connAttempts > 0 {
		pg.Pool, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
		if err != nil {
			return nil, fmt.Errorf("postgres - NewPostgres - pgxpool.NewWithConfig: %w", err)
		}

		if err = pg.Pool.Ping(context.Background()); err == nil {
			break
		}

		log.Infof("Postgres is trying to connect, attempts left: %d", pg.connAttempts)
		pg.connAttempts--
		time.Sleep(pg.connTimeout)
	}

	if err != nil {
		return nil, fmt.Errorf("postgres - NewPostgres - connAttempts == 0: %w", err)
	}

	return pg, nil
}

func (pg *Postgres) Close() {
	if pg.Pool != nil {
		pg.Pool.Close()
	}
}

// GetQueryRunner возвращает объект для выполнения запросов: если в контексте есть транзакция,
// а иначе использует пул соединений.
func (pg *Postgres) GetQueryRunner(ctx context.Context) QueryRunner {
	if tx, ok := extractTx(ctx); ok {
		return tx
	}
	return pg.Pool
}

// WithinTransaction выполняет функцию fn в рамках транзакции.
func (pg *Postgres) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := pg.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("postgres - WithinTransaction - Begin: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err = fn(injectTx(ctx, tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
