package service

import (
	"OilStore/internal/domain"
	"OilStore/internal/logger"
	"OilStore/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
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
	oilsAllKey    = "oils:all"
	oilByIdPr     = "oils:"
	oilAbovePrice = "oils:price:above:"
	cacheTTL      = 5 * time.Minute
)

func (s *OilService) invalidatePrefix(ctx context.Context, prefix string) {
	iter := s.redisCli.Scan(ctx, 0, prefix+"*", 100).Iterator()
	for iter.Next(ctx) {
		s.redisCli.Del(ctx, iter.Val())

	}
}

func (s *OilService) AddOil(ctx context.Context, oil domain.AddOilDomain) (int, error) {
	if oil.Name == "" {
		logger.Log.Warn("Oil name can not be empty")
		return 0, fmt.Errorf("name can not be empty")
	}

	if oil.Price < 0 {
		logger.Log.Warn("Price can not be <0", zap.Int("price", oil.Price))
		return 0, fmt.Errorf("price can not be <0")
	}

	if oil.Visc == "" {
		logger.Log.Warn("Visc can not be empty")
		return 0, fmt.Errorf("visc can not be empty")
	}

	oilDB := models.Oil{
		Name:  oil.Name,
		Visc:  oil.Visc,
		Price: oil.Price,
	}
	id, err := s.oilRepo.AddOil(ctx, oilDB)
	if err != nil {

		logger.Log.Error("DB error. New oil was not added in DB", zap.String("name", oil.Name), zap.Error(err))
		return 0, err
	}
	logger.Log.Info("Success: new oil was added in DB",
		zap.Int("id", id), zap.String("name", oil.Name),
		zap.String("visc", oil.Visc),
		zap.Int("price", oil.Price))

	s.redisCli.Del(ctx, oilsAllKey)
	s.invalidatePrefix(ctx, "oils:price:above:")
	return id, err

}

func (s *OilService) DeleteOilById(ctx context.Context, id int) error {
	err := s.oilRepo.DeleteOilById(ctx, id)
	if err != nil {
		logger.Log.Error("DB error! Oil was not deleted from DB", zap.Int("id", id), zap.Error(err))
		return err
	}
	logger.Log.Info("Success: oil was deleted from DB", zap.Int("id", id))
	s.redisCli.Del(ctx, oilsAllKey)
	s.redisCli.Del(ctx, oilByIdPr+strconv.Itoa(id))
	s.invalidatePrefix(ctx, "oils:price:above:")
	return nil
}

func (s *OilService) FullUpdateOil(ctx context.Context, oil domain.OilDomain, id int) (domain.OilDomain, error) {

	oilDB := models.Oil{
		Name:  oil.Name,
		Visc:  oil.Visc,
		Price: oil.Price,
	}
	retOil, err := s.oilRepo.FullUpdateOil(ctx, oilDB, id)
	if err == nil {
		s.redisCli.Del(ctx, oilsAllKey)
		s.redisCli.Del(ctx, oilByIdPr+strconv.Itoa(id))
		s.invalidatePrefix(ctx, "oils:price:above:")
		logger.Log.Info("Oil was successfully updated! ", zap.Int("id", retOil.Id), zap.String("name", retOil.Name))

		oilSV := domain.OilDomain{
			Id:    retOil.Id,
			Name:  retOil.Name,
			Visc:  retOil.Visc,
			Price: retOil.Price,
		}
		return oilSV, err
	}
	logger.Log.Error("DB error! Oil was not updated!", zap.Int("id", retOil.Id), zap.String("name", retOil.Name), zap.Error(err))
	return domain.OilDomain{}, err
}

func (s *OilService) GetMinMaxOil(ctx context.Context, min, max int) ([]domain.OilDomain, error) {
	if min < 0 {
		logger.Log.Warn("Minimum price can not be <0", zap.Int("min", min))
		return nil, fmt.Errorf("minimum price can not be <0")
	}

	oils, errRep := s.oilRepo.GetMinMaxOil(ctx, min, max)

	oilsSV := make([]domain.OilDomain, len(oils))

	for i, v := range oils {
		oilsSV[i] = domain.OilDomain{
			Id:    v.Id,
			Name:  v.Name,
			Visc:  v.Visc,
			Price: v.Price,
		}

	}

	if errRep != nil {
		logger.Log.Error("DB error! oils was not reiceved from DB", zap.Error(errRep))
		return nil, errRep
	}
	if len(oils) == 0 {

	}
	logger.Log.Info("Successfully fetched oil data from the database", zap.Int("count", len(oils)))
	return oilsSV, nil

}
func (s *OilService) GetByVisc(ctx context.Context, visc string) ([]domain.OilDomain, error) {
	if visc == "" {
		logger.Log.Warn("Visc can not be empty!", zap.String("visc", visc))
		return nil, fmt.Errorf("visc can not be empty")
	}

	oils, err := s.oilRepo.GetByVisc(ctx, visc)

	if err != nil {
		logger.Log.Error("DB error! oils was not founded in DB", zap.Error(err))
		return nil, err
	}

	oilsSV := make([]domain.OilDomain, len(oils))

	for i, v := range oils {
		oilsSV[i] = domain.OilDomain{
			Id:    v.Id,
			Name:  v.Name,
			Visc:  v.Visc,
			Price: v.Price,
		}
	}

	return oilsSV, err

}

func (s *OilService) GetAllOils(ctx context.Context) ([]domain.OilDomain, error) {

	cached, err := s.redisCli.Get(ctx, oilsAllKey).Result()
	if err == nil {
		var oils []domain.OilDomain
		if err := json.Unmarshal([]byte(cached), &oils); err == nil {
			return oils, nil
		}

		logger.Log.Warn("Failed to unmarshal cached oils, falling back to DB")
	} else if err != redis.Nil {
		logger.Log.Warn("Redis unavailble, falling back to DB", zap.Error(err))
	}

	oils, err := s.oilRepo.GetAllOils(ctx)

	if err != nil {
		logger.Log.Error("failed to fetch oils from DB", zap.Error(err))
		return nil, err
	}

	oilsSV := make([]domain.OilDomain, len(oils))

	for i, v := range oils {
		oilsSV[i] = domain.OilDomain{
			Id:    v.Id,
			Name:  v.Name,
			Visc:  v.Visc,
			Price: v.Price,
		}
	}

	redData, errMarshal := json.Marshal(oilsSV)
	if errMarshal == nil {
		s.redisCli.Set(ctx, oilsAllKey, redData, cacheTTL)
		logger.Log.Debug("put data from postgres to Redis")
	} else {
		logger.Log.Warn("Failed to marshal slice for Redis", zap.Error(err))
	}
	logger.Log.Info("oils fetched", zap.Int("count", len(oils)))
	return oilsSV, nil
}

func (s *OilService) GetOilById(ctx context.Context, id int) (domain.OilDomain, error) {
	redisKey := oilByIdPr + strconv.Itoa(id)
	cached, err := s.redisCli.Get(ctx, redisKey).Result()
	if err == nil {
		var cacheOil domain.OilDomain
		logger.Log.Info("Get data from Redis")
		if err := json.Unmarshal([]byte(cached), &cacheOil); err == nil {
			return cacheOil, nil
		}
	}

	oil, err := s.oilRepo.GetOilById(ctx, id)

	if err != nil {
		logger.Log.Error("DB error!", zap.Int("id", id), zap.Error(err))
		return domain.OilDomain{}, err
	}
	oilSV := domain.OilDomain{
		Id:    oil.Id,
		Name:  oil.Name,
		Visc:  oil.Visc,
		Price: oil.Price,
	}

	reqData, err := json.Marshal(oilSV)
	if err == nil {
		logger.Log.Debug("Put data to Redis")
		s.redisCli.Set(ctx, redisKey, reqData, cacheTTL)
	}
	logger.Log.Info("success get oil by id!", zap.Int("id", id))
	return oilSV, err
}

func (s *OilService) GetOilsAbovePrice(ctx context.Context, price int) ([]domain.OilDomain, error) {

	redisKey := oilAbovePrice + strconv.Itoa(price)

	if price < 0 {
		logger.Log.Warn("price can not be a lower, than 0", zap.Int("price", price))
		return nil, fmt.Errorf("price can not be low a 0 (zero)")
	}

	cached, errRedis := s.redisCli.Get(ctx, redisKey).Result()
	if errRedis == nil {
		logger.Log.Debug("Get data from Redis")
		var oilsSV []domain.OilDomain
		if errUnmarshalRedis := json.Unmarshal([]byte(cached), &oilsSV); errUnmarshalRedis == nil {
			return oilsSV, nil
		}
		logger.Log.Warn("Can not unmarshal data", zap.Error(errRedis))

	} else if errRedis != redis.Nil {
		logger.Log.Warn("redis unavaible", zap.Error(errRedis))
	}

	oils, errOilsRepo := s.oilRepo.GetOilsAbovePrice(ctx, price)
	if errOilsRepo != nil {
		logger.Log.Error("DB error!", zap.Error(errOilsRepo))
		return nil, errOilsRepo
	}

	oilsSV := make([]domain.OilDomain, len(oils))

	for i, v := range oils {
		oilsSV[i] = domain.OilDomain{
			Id:    v.Id,
			Name:  v.Name,
			Visc:  v.Visc,
			Price: v.Price,
		}
	}

	newOilsRep, errMarshal := json.Marshal(oilsSV)
	if errMarshal != nil {
		logger.Log.Warn("Can't marshal data from DB", zap.Error(errMarshal))

	} else {
		s.redisCli.Set(ctx, redisKey, newOilsRep, cacheTTL)
	}

	logger.Log.Info("Success to get data from postgreSQL", zap.Int("price", price), zap.Int("count", len(oils)))
	return oilsSV, nil
}
