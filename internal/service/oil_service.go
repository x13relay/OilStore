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
	//1.храним полные данные масла в json
	oilDataKeyPr = "oils:data:"
	//2. Кэш списков и фильтров. Храним только слайсы айдишников.
	oilsListAll     = "oils:list:all"
	oilsViscKeyPr   = "oils:list:visc:"
	oilsMinMaxKeyPr = "oils:list:price:range:"
	oilsAbovePrice  = "oils:price:above:"
	oilFilterTTL    = 5 * time.Minute
	oilDataTTL      = 24 * time.Hour
)

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

	return id, err

}

func (s *OilService) DeleteOilById(ctx context.Context, id int) error {
	err := s.oilRepo.DeleteOilById(ctx, id)
	if err != nil {
		logger.Log.Error("DB error! Oil was not deleted from DB", zap.Int("id", id), zap.Error(err))
		return err
	}
	logger.Log.Info("Success: oil was deleted from DB", zap.Int("id", id))
	s.redisCli.Del(ctx, oilDataKeyPr+strconv.Itoa(id))
	logger.Log.Info("oil was success deleted from redis.", zap.Int("ID", id))
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
		logger.Log.Info("Oil was successfully updated! ", zap.Int("id", retOil.Id), zap.String("name", retOil.Name))
		s.redisCli.Del(ctx, oilDataKeyPr+strconv.Itoa(id))
		logger.Log.Info("updated oil was success deleted from Redis", zap.Int("ID", id))

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

	dinamicRedisKey := oilsMinMaxKeyPr + strconv.Itoa(min) + "-" + strconv.Itoa(max)

	cachedOil, errCachedOil := s.redisCli.Get(ctx, dinamicRedisKey).Result()
	var oilIDs []int
	if errCachedOil == nil {

		errJsonUnmarshal := json.Unmarshal([]byte(cachedOil), &oilIDs)
		if errJsonUnmarshal != nil {
			logger.Log.Error("error unmarshal slice of IDs", zap.Error(errJsonUnmarshal))
		}
	}
	if len(oilIDs) == 0 || errCachedOil == redis.Nil {
		oilsDB, errDB := s.oilRepo.GetMinMaxOil(ctx, min, max)
		if errDB != nil {
			logger.Log.Error("Database error", zap.Error(errDB))
			return nil, errDB
		}
		if len(oilsDB) == 0 {
			return []domain.OilDomain{}, nil
		}
		oilIDs = make([]int, len(oilsDB))

		for i, v := range oilsDB {
			oilIDs[i] = v.Id
		}

		sliceIDs, errIDs := json.Marshal(oilIDs)
		if errIDs == nil {
			s.redisCli.Set(ctx, dinamicRedisKey, sliceIDs, oilFilterTTL).Result()
			logger.Log.Info("put IDs to Redis", zap.String("IDs", string(sliceIDs)))
		} else {
			logger.Log.Error("error marshal slice of IDs", zap.Error(errIDs))
		}

	}

	mgetKeys := make([]string, len(oilIDs))

	for i, id := range oilIDs {
		mgetKeys[i] = oilDataKeyPr + strconv.Itoa(id)
	}
	mgetSlice, errMget := s.redisCli.MGet(ctx, mgetKeys...).Result()
	if errMget != nil {
		logger.Log.Error("MGET error", zap.Error(errMget))
	}
	resultSlice := make([]domain.OilDomain, 0, len(oilIDs))

	for i, item := range mgetSlice {
		currentID := oilIDs[i]
		if item != nil {
			var oilReq domain.OilDomain
			errUnmarshal := json.Unmarshal([]byte(item.(string)), &oilReq)
			if errUnmarshal == nil {
				resultSlice = append(resultSlice, oilReq)
				continue
			}
			logger.Log.Warn("Failed to unmarshal single oil data from Redis", zap.Error(errUnmarshal))

		}
		logger.Log.Info("single product cache miss. Get product from DB", zap.Int("ID", currentID))
		oilDB, errDB := s.oilRepo.GetOilById(ctx, currentID)
		if errDB != nil {
			logger.Log.Error("error single product get from DB", zap.Error(errDB))
			continue
		}
		oilSV := domain.OilDomain{
			Id:    oilDB.Id,
			Name:  oilDB.Name,
			Visc:  oilDB.Visc,
			Price: oilDB.Price,
		}
		resultSlice = append(resultSlice, oilSV)
		oilData, errData := json.Marshal(oilSV)
		if errData == nil {
			individualKey := oilDataKeyPr + strconv.Itoa(currentID)
			s.redisCli.Set(ctx, individualKey, oilData, oilDataTTL)
			logger.Log.Info("put single product to Redis", zap.String("oil", string(oilData)))
		} else {
			logger.Log.Warn("error marchal oildata", zap.Error(errData))
		}
	}
	logger.Log.Info("oils successfully fetched via two-lvl cache", zap.Int("count", len(resultSlice)))
	return resultSlice, nil
}

func (s *OilService) GetByVisc(ctx context.Context, visc string) ([]domain.OilDomain, error) {
	if visc == "" {
		logger.Log.Warn("Visc can not be empty!", zap.String("visc", visc))
		return nil, fmt.Errorf("visc can not be empty")
	}
	dinamicRedisKey := oilsViscKeyPr + visc

	cachedVisc, errCachedVisc := s.redisCli.Get(ctx, dinamicRedisKey).Result()
	var oilsIDs []int
	if errCachedVisc == nil {
		json.Unmarshal([]byte(cachedVisc), &oilsIDs)
	}

	if errCachedVisc == redis.Nil || len(oilsIDs) == 0 {
		oilsDB, errDB := s.oilRepo.GetByVisc(ctx, visc)
		if errDB != nil {
			logger.Log.Error("DB error! oils was not founded in DB", zap.Error(errDB))
			return nil, errDB
		}
		if len(oilsDB) == 0 {
			return []domain.OilDomain{}, nil
		}

		oilsIDs = make([]int, len(oilsDB))

		for i, v := range oilsDB {
			oilsIDs[i] = v.Id
		}

		sliceIDs, errIDs := json.Marshal(oilsIDs)
		if errIDs == nil {
			s.redisCli.Set(ctx, dinamicRedisKey, sliceIDs, oilFilterTTL).Result()
			logger.Log.Info("put IDs slice to Redis success", zap.String("ids", string(sliceIDs)))
		} else {
			logger.Log.Error("error marshal ids to slice", zap.Error(errIDs))
		}
	}

	mgetKeys := make([]string, len(oilsIDs))

	for i, id := range oilsIDs {
		mgetKeys[i] = oilDataKeyPr + strconv.Itoa(id)
	}
	mgetSlice, errorMget := s.redisCli.MGet(ctx, mgetKeys...).Result()
	if errorMget != nil {
		logger.Log.Error("MGET error", zap.Error(errorMget))
	}
	resultSlice := make([]domain.OilDomain, 0, len(oilsIDs))

	for i, item := range mgetSlice {
		currentId := oilsIDs[i]
		if item != nil {
			var oil domain.OilDomain
			if err := json.Unmarshal([]byte(item.(string)), &oil); err == nil {
				resultSlice = append(resultSlice, oil)
				continue
			}
			logger.Log.Warn("Failed to unmarshal single oil data from Redis", zap.Int("id", currentId))
		}

		logger.Log.Info("single product cache miss, fetching row from DB", zap.Int("id", currentId))
		DBoil, errDB := s.oilRepo.GetOilById(ctx, currentId)
		if errDB != nil {
			logger.Log.Error("DB error while fetching single oil", zap.Error(errDB))
			continue
		}

		oilSV := domain.OilDomain{
			Id:    DBoil.Id,
			Name:  DBoil.Name,
			Visc:  DBoil.Visc,
			Price: DBoil.Price,
		}
		resultSlice = append(resultSlice, oilSV)
		reqData, errMarshal := json.Marshal(oilSV)
		if errMarshal == nil {
			individualKey := oilDataKeyPr + strconv.Itoa(currentId)
			s.redisCli.Set(ctx, individualKey, reqData, oilDataTTL)
			logger.Log.Info("put single product row to Redis", zap.String("key", individualKey))
		} else {
			logger.Log.Warn("Failed to marshal single oil", zap.Error(errMarshal))
		}

	}

	logger.Log.Info("oils successfully fetched via two-lvl cache", zap.Int("count", len(resultSlice)))
	return resultSlice, nil
}

func (s *OilService) GetAllOils(ctx context.Context) ([]domain.OilDomain, error) {
	var oilIds []int
	cachedIDs, errCachedIDs := s.redisCli.Get(ctx, oilsListAll).Result()

	if errCachedIDs == nil {
		logger.Log.Info("Get lists of IDs from Redis", zap.String("key", oilsListAll))
		_ = json.Unmarshal([]byte(cachedIDs), &oilIds)
	} else if errCachedIDs != redis.Nil {
		logger.Log.Warn("Redis unavaible, falling back to DB", zap.Error(errCachedIDs))
	}

	if errCachedIDs == redis.Nil || len(oilIds) == 0 {
		logger.Log.Info("cache miss for lists of IDs, fetching from postgres.")
		oilsDB, errDB := s.oilRepo.GetAllOils(ctx)
		if errDB != nil {
			logger.Log.Error("Database error", zap.Error(errDB))
			return nil, errDB
		}

		oilIds = make([]int, len(oilsDB))

		for i, v := range oilsDB {
			oilIds[i] = v.Id
		}

		reqData, errMarshal := json.Marshal(oilIds)
		if errMarshal == nil {
			s.redisCli.Set(ctx, oilsListAll, reqData, oilFilterTTL)
			logger.Log.Info("Put list of IDs to Redis", zap.String("key", oilsListAll))
		} else {
			logger.Log.Warn("Failed to marshal oil IDs slice for Redis", zap.Error(errMarshal))
		}
	}

	if len(oilIds) == 0 {
		return []domain.OilDomain{}, nil
	}

	mgetKeys := make([]string, len(oilIds))
	for i, id := range oilIds {
		mgetKeys[i] = oilDataKeyPr + strconv.Itoa(id)
	}
	mgetResults, errMget := s.redisCli.MGet(ctx, mgetKeys...).Result()
	if errMget != nil {
		logger.Log.Warn("Mget failed, try to recover missing item from DB", zap.Error(errMget))

	}
	resultOils := make([]domain.OilDomain, 0, len(oilIds))

	for i, item := range mgetResults {
		currentID := oilIds[i]
		if item != nil {
			var oil domain.OilDomain
			if err := json.Unmarshal([]byte(item.(string)), &oil); err == nil {
				resultOils = append(resultOils, oil)
				continue
			}
			logger.Log.Warn("Failed to unmarshal single oil data from Redis", zap.Int("id", currentID))

		}
		logger.Log.Info("single product cache miss, fetchingrow from DB", zap.Int("id", currentID))
		DBoil, errDB := s.oilRepo.GetOilById(ctx, currentID)
		if errDB != nil {
			logger.Log.Error("DB error while fetching simgle oil", zap.Error(errDB))
			continue
		}
		oilSV := domain.OilDomain{
			Id:    DBoil.Id,
			Name:  DBoil.Name,
			Visc:  DBoil.Visc,
			Price: DBoil.Price,
		}

		resultOils = append(resultOils, oilSV)

		reqData, errMarshal := json.Marshal(oilSV)
		if errMarshal == nil {
			individualKey := oilDataKeyPr + strconv.Itoa(currentID)
			s.redisCli.Set(ctx, individualKey, reqData, oilDataTTL)
			logger.Log.Info("put single product row to Redis", zap.String("key", individualKey))
		} else {
			logger.Log.Warn("Failed to marshal single oil", zap.Error(errMarshal))
		}
	}
	logger.Log.Info("oils successfully fetched via two-lvl cache", zap.Int("count", len(resultOils)))
	return resultOils, nil
}

func (s *OilService) GetOilById(ctx context.Context, id int) (domain.OilDomain, error) {
	redisKey := oilDataKeyPr + strconv.Itoa(id)
	cached, err := s.redisCli.Get(ctx, redisKey).Result()
	if err == nil {
		var cacheOil domain.OilDomain
		logger.Log.Info("Get data from Redis")
		if err := json.Unmarshal([]byte(cached), &cacheOil); err == nil {
			return cacheOil, nil
		}
	}

	oilDB, errDB := s.oilRepo.GetOilById(ctx, id)

	if errDB != nil {
		logger.Log.Error("DB error!", zap.Int("id", id), zap.Error(errDB))
		return domain.OilDomain{}, errDB
	}
	oilSV := domain.OilDomain{
		Id:    oilDB.Id,
		Name:  oilDB.Name,
		Visc:  oilDB.Visc,
		Price: oilDB.Price,
	}

	reqData, err := json.Marshal(oilSV)
	if err == nil {
		logger.Log.Debug("Put data to Redis")
		s.redisCli.Set(ctx, redisKey, reqData, oilDataTTL)
	}
	logger.Log.Info("success get oil by id from DB", zap.Int("id", id))
	return oilSV, err
}

func (s *OilService) GetOilsAbovePrice(ctx context.Context, price int) ([]domain.OilDomain, error) {

	if price < 0 {
		logger.Log.Warn("price can not be a lower, than 0", zap.Int("price", price))
		return nil, fmt.Errorf("price can not be low a 0 (zero)")
	}
	var oilsIDs []int

	dynamicRedisKey := oilsAbovePrice + strconv.Itoa(price)

	cached, errRedis := s.redisCli.Get(ctx, dynamicRedisKey).Result()
	if errRedis == nil {
		logger.Log.Debug("Get data from Redis")

		_ = json.Unmarshal([]byte(cached), &oilsIDs)

	}

	if errRedis == redis.Nil || len(oilsIDs) == 0 {
		logger.Log.Warn("miss cache IDs from radis. Get IDs from DB")
		oilDB, errDB := s.oilRepo.GetOilsAbovePrice(ctx, price)
		if errDB != nil {
			logger.Log.Error("DB error", zap.Error(errDB))
			return nil, errDB
		}

		if len(oilDB) == 0 {
			return []domain.OilDomain{}, nil
		}

		oilsIDs = make([]int, len(oilDB))

		for i, v := range oilDB {
			oilsIDs[i] = v.Id
		}
		sliceIDs, errIDs := json.Marshal(oilsIDs)
		if errIDs == nil {
			s.redisCli.Set(ctx, dynamicRedisKey, sliceIDs, oilFilterTTL)
			logger.Log.Info("put IDs slice to Redis success", zap.String("key", dynamicRedisKey))

		} else {
			logger.Log.Error("error marshal IDs slice", zap.Error(errIDs))
		}

	}

	mgetKeys := make([]string, len(oilsIDs))

	for i, id := range oilsIDs {
		mgetKeys[i] = oilDataKeyPr + strconv.Itoa(id)
	}
	sliceMget, errMget := s.redisCli.MGet(ctx, mgetKeys...).Result()
	if errMget != nil {
		logger.Log.Error("MGET error", zap.Error(errMget))
	}
	resultSlice := make([]domain.OilDomain, 0, len(oilsIDs))

	for i, item := range sliceMget {
		currentId := oilsIDs[i]
		if item != nil {
			var oil domain.OilDomain
			if err := json.Unmarshal([]byte(item.(string)), &oil); err == nil {
				resultSlice = append(resultSlice, oil)
				continue
			}
			logger.Log.Warn("Failed to unmarshal single oil data from Redis", zap.Int("id", currentId))
		}
		logger.Log.Info("single product cache miss, fetching row from DB", zap.Int("id", currentId))
		oilDB, errDB := s.oilRepo.GetOilById(ctx, currentId)
		if errDB != nil {
			logger.Log.Error("DB error while fetching single oil", zap.Error(errDB))
			continue
		}

		oilSV := domain.OilDomain{
			Id:    oilDB.Id,
			Name:  oilDB.Name,
			Visc:  oilDB.Visc,
			Price: oilDB.Price,
		}
		resultSlice = append(resultSlice, oilSV)

		reqData, errData := json.Marshal(oilSV)
		if errData == nil {
			individualKey := oilDataKeyPr + strconv.Itoa(currentId)
			s.redisCli.Set(ctx, individualKey, reqData, oilDataTTL)
			logger.Log.Info("put single product row to Redis", zap.String("key", individualKey))
		} else {
			logger.Log.Error("marshal error", zap.Error(errData))
		}
	}
	return resultSlice, nil
}
