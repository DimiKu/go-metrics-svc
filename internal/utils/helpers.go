package utils

import (
	"context"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"log"
	"time"
)

// CommitOrRollback ф-я для автматического роллбека при ошибке или коммита при отсутствии ошибки
func CommitOrRollback(tx pgx.Tx, err error, ctx context.Context) {
	if err != nil {
		tx.Rollback(ctx)
	} else {
		tx.Commit(ctx)
	}
}

// ф-я для првоерки ошибков ПГ
func isRetryable(err error) bool {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		switch pgErr.Code {
		case pgerrcode.ConnectionException, pgerrcode.ConnectionDoesNotExist, pgerrcode.ConnectionFailure, pgerrcode.CannotConnectNow:
			return true
		}
	}
	return false
}

// RetryableQuery ф-я для выполнения запросов на чтение с ретраем
func RetryableQuery(ctx context.Context, pool *pgxpool.Pool, log *zap.SugaredLogger, query string, args ...interface{}) (pgx.Rows, error) {
	var row pgx.Rows
	var err error

	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer CommitOrRollback(tx, err, ctx)

	for i := 0; i < 3; i++ {
		log.Debugf("Im do query: %s, time: %d", query, i)
		row, err = tx.Query(ctx, query, args...)
		if err != nil {
			if isRetryable(err) {
				log.Debugf("Retrying query due to error: %v\n", err)
				time.Sleep(time.Duration(i) * time.Second)
				log.Debugf("Im do query: %s", query)
				continue
			}

		}
		return row, nil
	}
	return nil, err
}

// RetryableExec ф-я для выполнения запросов на запись с ретраем
func RetryableExec(ctx context.Context, pool *pgxpool.Pool, command string, args ...interface{}) (pgconn.CommandTag, error) {
	var tag pgconn.CommandTag
	var err error

	tx, err := pool.Begin(ctx)
	if err != nil {
		return tag, err
	}

	defer CommitOrRollback(tx, err, ctx)

	for i := 0; i < 3; i++ {
		tag, err = tx.Exec(ctx, command, args...)
		if err != nil {
			if isRetryable(err) {
				log.Printf("Retrying exec due to error: %v\n", err)
				time.Sleep(time.Duration(i) * time.Second)
				continue
			}
			return tag, err
		}
		return tag, nil
	}
	return tag, err
}
