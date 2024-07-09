package reporedis

import (
	"bytes"
	"conference/internal/sound"
	"conference/internal/sound/soundwav"
	"context"
	"encoding/gob"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// TODO check if I need to use there mutex
type RepositoryRedis struct {
	ctx    context.Context
	client *redis.Client
	logger *zap.Logger
}

// may be slow because of reflect inside
// add concurancy later

func NewRepo(ctx context.Context, addr string, lr *zap.Logger) *RepositoryRedis {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})
	return &RepositoryRedis{ctx: ctx, client: client, logger: lr.With(zap.String("app", "redisRepo"))}
}

func (repo *RepositoryRedis) SetSound(k string, v *sound.Sound, tExp time.Duration) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(*(*v).(*soundwav.SoundWav)); err != nil {
		repo.logger.Error("Error occured during encoding sound", zap.String("key", k), zap.Error(err))
		return err
	}

	if err := repo.client.Set(repo.ctx, k, buf.String(), tExp).Err(); err != nil {
		repo.logger.Error("Error occured during set to redis", zap.String("key", k), zap.Error(err))
		return err
	}

	return nil
}

// TODO Check for redis.nil error
// TODO check if I really need such conversations between interface and it's impl
func (repo *RepositoryRedis) GetSound(k string) (bool, *sound.Sound, error) {
	res, err := repo.client.Get(repo.ctx, k).Bytes()
	repo.logger.Info("GetSound", zap.String("key", k), zap.Error(err))
	if err == redis.Nil {
		return false, nil, nil
	}
	if err != nil {
		repo.logger.Error("Error occured during reading sound from redis", zap.String("key", k), zap.Error(err))
		return false, nil, err
	}

	var buf *bytes.Buffer = bytes.NewBuffer(res)
	enc := gob.NewDecoder(buf)

	var sw soundwav.SoundWav
	repo.logger.Info("GetSound", zap.String("key", k), zap.Error(err))
	if err := enc.Decode(&sw); err != nil {
		return false, nil, err
	}

	s := sound.Sound(&sw)
	return true, &s, nil
}

func (repo *RepositoryRedis) GetDelSound(k string) (bool, *sound.Sound, error) {
	res, err := repo.client.GetDel(repo.ctx, k).Bytes()
	if err == redis.Nil {
		return false, nil, nil
	}
	if err != nil {
		repo.logger.Error("Error occured during reading sound from redis", zap.String("key", k), zap.Error(err))
		return false, nil, err
	}

	var buf *bytes.Buffer = bytes.NewBuffer(res)
	enc := gob.NewDecoder(buf)

	var sw soundwav.SoundWav
	repo.logger.Info("GetSound", zap.String("key", k), zap.Error(err))
	if err := enc.Decode(&sw); err != nil {
		return false, nil, err
	}

	s := sound.Sound(&sw)
	return true, &s, nil
}
