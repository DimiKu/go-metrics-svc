package storage

import (
	"context"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type DBStorage struct {
	conn *pgx.Conn

	log *zap.SugaredLogger
}

func NewDBStorage(conn *pgx.Conn, log *zap.SugaredLogger) *DBStorage {
	return &DBStorage{
		conn: conn,
		log:  log,
	}
}

func (d *DBStorage) DBPing(ctx context.Context) (bool, error) {
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
