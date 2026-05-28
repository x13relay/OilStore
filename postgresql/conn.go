package postgresql

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5"
)

type OilConn struct {
	conn *pgx.Conn
}

func NewOilConn(conn *pgx.Conn) *OilConn {
	return &OilConn{conn: conn}
}

func ConnectionBD_oil(ctx context.Context) (*pgx.Conn, error) {
	conStr := os.Getenv("CONSTR")
	return pgx.Connect(ctx, conStr)
}
