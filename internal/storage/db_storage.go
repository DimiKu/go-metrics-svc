package storage

import (
	"context"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type DbStorage struct {
	conn *pgx.Conn

	log *zap.SugaredLogger
}

func NewDbStorage(conn *pgx.Conn, log *zap.SugaredLogger) *DbStorage {
	return &DbStorage{
		conn: conn,
		log:  log,
	}
}

func (d *DbStorage) DbPing(ctx context.Context) (bool, error) {
	var result int
	if d.conn == nil {
		return false, nil
	}
	err := d.conn.QueryRow(ctx, PingQuery).Scan(&result)
	if err != nil {
		return false, err
	}
	if result == 1 {
		return true, nil
	}

	return false, err
}
