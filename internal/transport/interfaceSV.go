package transport

import (
	"OilStore/internal/domain"
	"context"
)

type OilService interface {
	AddOil(ctx context.Context, oil domain.AddOilDomain) (int, error)
	DeleteOilById(ctx context.Context, id int) error
	FullUpdateOil(ctx context.Context, oil domain.OilDomain, id int) (domain.OilDomain, error)
	GetMinMaxOil(ctx context.Context, min, max int) ([]domain.OilDomain, error)
	GetByVisc(ctx context.Context, visc string) ([]domain.OilDomain, error)
	GetAllOils(ctx context.Context) ([]domain.OilDomain, error)
	GetOilById(ctx context.Context, id int) (domain.OilDomain, error)
	GetOilsAbovePrice(ctx context.Context, price int) ([]domain.OilDomain, error)
}
