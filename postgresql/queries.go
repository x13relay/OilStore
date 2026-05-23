package postgresql

import (
	"OilStore/models"
	"context"
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
	tableAddOilSQL := `INSERT INTO oils 
(name,visc,price)
VALUES($1,$2,$3)
RETURNING id
`
	var oId int

	err := oc.conn.QueryRow(ctx, tableAddOilSQL, oil.Name, oil.Visc, oil.Price).Scan(&oId)
	return oId, err
}
