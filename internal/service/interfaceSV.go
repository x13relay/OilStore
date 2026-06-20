package service

import (
	"OilStore/internal/models"
	"context"
)

type OilService interface {
	AddOil(ctx context.Context, oil models.Oil) (int, error)
	DeleteOilById(ctx context.Context, id int) error
	FullUpdateOil(ctx context.Context, oil models.Oil, id int) (models.Oil, error)
	GetMinMaxOil(ctx context.Context, min, max int) ([]models.Oil, error)
	GetByVisc(ctx context.Context, visc string) ([]models.Oil, error)
	GetAllOils(ctx context.Context) ([]models.Oil, error)
	GetOilById(ctx context.Context, id int) (models.Oil, error)
	GetOilsAbovePrice(ctx context.Context, price int) ([]models.Oil, error)
}
