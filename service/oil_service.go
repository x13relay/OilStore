package service

import (
	"OilStore/logger"
	"OilStore/models"
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

func (s *OilService) AddOil(ctx context.Context, oil models.Oil) (int, error) {
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

	id, err := s.oilRepo.AddOil(ctx, oil)
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

func (s *OilService) FullUpdateOil(ctx context.Context, oil models.Oil, id int) (models.Oil, error) {

	retOil, err := s.oilRepo.FullUpdateOil(ctx, oil, id)
	if err == nil {
		s.redisCli.Del(ctx, oilsAllKey)
		s.redisCli.Del(ctx, oilByIdPr+strconv.Itoa(id))
		s.invalidatePrefix(ctx, "oils:price:above:")
		logger.Log.Info("Oil was successfully updated! ", zap.Int("id", retOil.Id), zap.String("name", retOil.Name))
		return retOil, err
	}
	logger.Log.Error("DB error! Oil was not updated!", zap.Int("id", retOil.Id), zap.String("name", retOil.Name), zap.Error(err))
	return models.Oil{}, err
}

func (s *OilService) GetMinMaxOil(ctx context.Context, min, max int) ([]models.Oil, error) {
	if min < 0 {
		logger.Log.Warn("Minimum price can not be <0", zap.Int("min", min))
		return nil, fmt.Errorf("minimum price can not be <0")
	}

	oils, errRep := s.oilRepo.GetMinMaxOil(ctx, min, max)
	if errRep != nil {
		logger.Log.Error("DB error! oils was not reiceved from DB", zap.Error(errRep))
		return nil, errRep
	}
	if len(oils) == 0 {

	}
	logger.Log.Info("Successfully fetched oil data from the database", zap.Int("count", len(oils)))
	return oils, nil

}
func (s *OilService) GetByVisc(ctx context.Context, visc string) ([]models.Oil, error) {
	if visc == "" {
		logger.Log.Warn("Visc can not be empty!", zap.String("visc", visc))
		return nil, fmt.Errorf("visc can not be empty")
	}

	oils, err := s.oilRepo.GetByVisc(ctx, visc)
	if err != nil {
		logger.Log.Error("DB error! oils was not founded in DB", zap.Error(err))
		return nil, err
	}

	return oils, err

}

func (s *OilService) GetAllOils(ctx context.Context) ([]models.Oil, error) {

	cached, err := s.redisCli.Get(ctx, oilsAllKey).Result()
	if err == nil {
		var oils []models.Oil
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

	redData, err := json.Marshal(oils)
	if err == nil {
		s.redisCli.Set(ctx, oilsAllKey, redData, cacheTTL)
		logger.Log.Debug("put data from postgres to Redis")
	} else {
		logger.Log.Warn("Failed to marshal slice for Redis", zap.Error(err))
	}
	logger.Log.Info("oils fetched", zap.Int("count", len(oils)))
	return oils, err
}

func (s *OilService) GetOilById(ctx context.Context, id int) (models.Oil, error) {
	redisKey := oilByIdPr + strconv.Itoa(id)
	cached, err := s.redisCli.Get(ctx, redisKey).Result()
	if err == nil {
		var oil models.Oil
		logger.Log.Warn("Get data from REdis")
		if err := json.Unmarshal([]byte(cached), &oil); err == nil {
			return oil, err
		}
	}
	oil, err := s.oilRepo.GetOilById(ctx, id)
	if err != nil {
		logger.Log.Error("DB error!", zap.Int("id", id), zap.Error(err))
		return models.Oil{}, err
	}
	reqData, err := json.Marshal(oil)
	if err == nil {
		logger.Log.Debug("Put data to Redis")
		s.redisCli.Set(ctx, redisKey, reqData, cacheTTL)
	}
	logger.Log.Info("success get oil by id!", zap.Int("id", id))
	return oil, err
}

func (s *OilService) GetOilsAbovePrice(ctx context.Context, price int) ([]models.Oil, error) {

	redisKey := oilAbovePrice + strconv.Itoa(price)

	if price < 0 {
		logger.Log.Warn("price can not be a lower, than 0", zap.Int("price", price))
		return nil, fmt.Errorf("price can not be low a 0 (zero)")
	}

	var newOils []models.Oil

	cached, errRedis := s.redisCli.Get(ctx, redisKey).Result()
	if errRedis == nil {
		logger.Log.Debug("Get data from Redis")
		if errUnmarshalRedis := json.Unmarshal([]byte(cached), &newOils); errUnmarshalRedis == nil {
			return newOils, nil
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

	newOilsRep, errMarshal := json.Marshal(oils)
	if errMarshal != nil {
		logger.Log.Warn("Can't marshal data from DB", zap.Error(errMarshal))

	} else {
		s.redisCli.Set(ctx, redisKey, newOilsRep, cacheTTL)
	}

	logger.Log.Info("Success to get data from postgreSQL", zap.Int("price", price), zap.Int("count", len(oils)))
	return oils, nil
}
