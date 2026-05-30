package service

import (
	"OilStore/models"
	"context"
	"fmt"
)

type OilService struct {
	oilRepo OilRepository
}

func NewOilService(oilRepo OilRepository) *OilService {
	return &OilService{
		oilRepo: oilRepo,
	}
}

func (s *OilService) AddOil(ctx context.Context, oil models.Oil) (int, error) {
	if oil.Name == "" {
		return 0, fmt.Errorf("name can not be empty")
	}

	if oil.Price < 0 {
		return 0, fmt.Errorf("price can not be <0")
	}

	if oil.Visc == "" {
		return 0, fmt.Errorf("visc can not be empty")
	}
	return s.oilRepo.AddOil(ctx, oil)
}

func (s *OilService) DeleteOilById(ctx context.Context, id int) error {
	return s.oilRepo.DeleteOilById(ctx, id)
}

func (s *OilService) FullUpdateOil(ctx context.Context, oil models.Oil, id int) (models.Oil, error) {

	return s.oilRepo.FullUpdateOil(ctx, oil, id)
}

func (s *OilService) GetMinMaxOil(ctx context.Context, min, max int) ([]models.Oil, error) {
	if min < 0 {
		return nil, fmt.Errorf("minimum price can not be <0")
	}
	return s.oilRepo.GetMinMaxOil(ctx, min, max)
}
func (s *OilService) GetByVisc(ctx context.Context, visc string) ([]models.Oil, error) {
	if visc == "" {
		return nil, fmt.Errorf("vusc can not be empty")
	}

	return s.oilRepo.GetByVisc(ctx, visc)
}

func (s *OilService) GetAllOils(ctx context.Context) ([]models.Oil, error) {
	return s.oilRepo.GetAllOils(ctx)
}
