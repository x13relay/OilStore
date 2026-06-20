package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type OilConn struct {
	conn *pgx.Conn
}

func NewOilConn(conn *pgx.Conn) *OilConn {
	return &OilConn{conn: conn}
}

func ConnectionDBPostgres(ctx context.Context, conStr string) (*pgx.Conn, error) {
	return pgx.Connect(ctx, conStr)
}
