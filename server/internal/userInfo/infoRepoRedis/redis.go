package infoRepoRedis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// TODO check if I need to use there mutex
type RepositoryRedis struct {
	ctx    context.Context
	client *redis.Client
	tExp   time.Duration
	logger *zap.Logger
}

// may be slow because of reflect inside
// TODO add concurancy later

func NewRepo(ctx context.Context, addr string, lr *zap.Logger) *RepositoryRedis {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})
	return &RepositoryRedis{ctx: ctx, client: client, logger: lr}
}

func (repo *RepositoryRedis) SetBitRate(k string, br int) error {
	return repo.client.Set(repo.ctx, k, br, repo.tExp).Err()
}

// TODO Check for redis.nil error
func (repo *RepositoryRedis) GetBitRate(k string) (bool, int, error) {
	br, err := repo.client.Get(repo.ctx, k).Int()
	if err == redis.Nil {
		return false, 0, nil
	}
	if err != nil {
		repo.logger.Error("Error occured during reading bitrate from redis", zap.String("key", k))
		return false, 0, err
	}

	return true, br, nil
}
