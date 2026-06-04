package service

import (
	"OilStore/models"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type OilService struct {
	oilRepo  OilRepository
	redisCli *redis.Client
}

func NewOilService(oilRepo OilRepository, rdb *redis.Client) *OilService {
	return &OilService{
		oilRepo:  oilRepo,
		redisCli: rdb,
	}
}

const (
	oilsAllKey = "oils:all"
	oilByIdPr  = "oils:"
	cacheTTL   = 5 * time.Minute
)

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

	id, err := s.oilRepo.AddOil(ctx, oil)
	if err == nil {
		s.redisCli.Del(ctx, oilsAllKey)
		return id, err
	}
	return 0, err
}

func (s *OilService) DeleteOilById(ctx context.Context, id int) error {
	s.redisCli.Del(ctx, oilsAllKey)
	s.redisCli.Del(ctx, oilByIdPr+strconv.Itoa(id))
	return s.oilRepo.DeleteOilById(ctx, id)
}

func (s *OilService) FullUpdateOil(ctx context.Context, oil models.Oil, id int) (models.Oil, error) {

	retOil, err := s.oilRepo.FullUpdateOil(ctx, oil, id)
	if err == nil {
		s.redisCli.Del(ctx, oilsAllKey)
		s.redisCli.Del(ctx, oilByIdPr+strconv.Itoa(id))
		return retOil, err
	}
	return models.Oil{}, err
}

func (s *OilService) GetMinMaxOil(ctx context.Context, min, max int) ([]models.Oil, error) {
	if min < 0 {
		return nil, fmt.Errorf("minimum price can not be <0")
	}
	return s.oilRepo.GetMinMaxOil(ctx, min, max)

}
func (s *OilService) GetByVisc(ctx context.Context, visc string) ([]models.Oil, error) {
	if visc == "" {
		return nil, fmt.Errorf("visc can not be empty")
	}

	return s.oilRepo.GetByVisc(ctx, visc)

}

func (s *OilService) GetAllOils(ctx context.Context) ([]models.Oil, error) {

	cached, err := s.redisCli.Get(ctx, oilsAllKey).Result()
	if err == nil {
		var oils []models.Oil
		if err := json.Unmarshal([]byte(cached), &oils); err == nil {
			return oils, nil
		}
	}

	oils, err := s.oilRepo.GetAllOils(ctx)
	if err != nil {
		return nil, err
	}
	redData, err := json.Marshal(oils)
	if err == nil {
		s.redisCli.Set(ctx, oilsAllKey, redData, cacheTTL)
	}
	return oils, err
}

func (s *OilService) GetOilById(ctx context.Context, id int) (models.Oil, error) {
	redisKey := oilByIdPr + strconv.Itoa(id)
	cached, err := s.redisCli.Get(ctx, redisKey).Result()
	if err == nil {
		var oil models.Oil
		if err := json.Unmarshal([]byte(cached), &oil); err == nil {
			return oil, err
		}
	}
	oil, err := s.oilRepo.GetOilById(ctx, id)
	if err != nil {
		return models.Oil{}, err
	}
	reqData, err := json.Marshal(oil)
	if err == nil {
		s.redisCli.Set(ctx, redisKey, reqData, cacheTTL)
	}
	return oil, err
}
