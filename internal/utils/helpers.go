package utils

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"log"
	"time"
)

func ArrStringContains(arr []string, str string) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}
	return false
}

func CommitOrRollback(tx pgx.Tx, err error, ctx context.Context) {
	if err != nil {
		tx.Rollback(ctx)
	} else {
		tx.Commit(ctx)
	}
}

func isRetryable(err error) bool {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		switch pgErr.Code {
		case "08000", "08001", "08003", "08004":
			return true
		}
	}
	return false
}

func RetryableQuery(ctx context.Context, pool *pgxpool.Pool, log *zap.SugaredLogger, query string, args ...interface{}) (pgx.Rows, error) {
	var row pgx.Rows
	var err error

	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer CommitOrRollback(tx, err, ctx)

	for i := 0; i < 3; i++ {
		log.Infof("Im do query: %s, time: %d", query, i)
		row, err = tx.Query(ctx, query, args...)
		if err != nil {
			if isRetryable(err) {
				log.Infof("Retrying query due to error: %v\n", err)
				time.Sleep(time.Duration(i) * time.Second)
				log.Infof("Im do query: %s", query)
				continue
			}

		}
		return row, nil
	}
	return nil, err
}

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
