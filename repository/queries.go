package repository

import (
	"OilStore/models"
	"context"

	"github.com/jackc/pgx/v5"
)

func (oc *OilConn) CreateTableOils(ctx context.Context) error {
	tableCr := `
CREATE TABLE IF NOT EXISTS oils 
(
id SERIAL PRIMARY KEY,
name VARCHAR (100) NOT NULL,
visc VARCHAR (100) NOT NULL,
price INTEGER NOT NULL
);
`

	_, err := oc.conn.Exec(ctx, tableCr)
	return err
}

func (oc *OilConn) AddOil(ctx context.Context, oil models.Oil) (int, error) {
	sQL := `INSERT INTO oils 
(name,visc,price)
VALUES($1,$2,$3)
RETURNING id
`
	var oId int

	err := oc.conn.QueryRow(ctx, sQL, oil.Name, oil.Visc, oil.Price).Scan(&oId)
	return oId, err
}

func (oc *OilConn) DeleteOilById(ctx context.Context, id int) error {
	sQL := `DELETE FROM oils WHERE id=$1`
	_, err := oc.conn.Exec(ctx, sQL, id)
	return err
}

func (oc *OilConn) FullUpdateOil(ctx context.Context, oil models.Oil, id int) (models.Oil, error) {
	sQL := `UPDATE oils
SET name=$1,visc=$2,price=$3
WHERE id=$4
RETURNING id,name,visc,price
`
	var roil models.Oil

	err := oc.conn.QueryRow(ctx, sQL, oil.Name, oil.Visc, oil.Price, id).Scan(&roil.Id, &roil.Name, &roil.Visc, &roil.Price)

	return roil, err
}

func (oc *OilConn) GetMinMaxOil(ctx context.Context, min, max int) ([]models.Oil, error) {
	sQl := `SELECT id,name,visc,price
	FROM oils WHERE price BETWEEN $1 AND $2`
	rows, err := oc.conn.Query(ctx, sQl, min, max)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[models.Oil])
}

func (oc *OilConn) GetByVisc(ctx context.Context, visc string) ([]models.Oil, error) {
	sQL := `SELECT id,name,visc,price
	FROM oils 
	WHERE REPLACE
	(LOWER(visc), '-', '') = REPLACE(LOWER($1), '-', '')`
	rows, err := oc.conn.Query(ctx, sQL, visc)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[models.Oil])
}
