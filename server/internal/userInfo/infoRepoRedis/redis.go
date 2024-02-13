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

func (repo *RepositoryRedis) SetId(k string, lId uint64) error {
	return repo.client.Set(repo.ctx, k, lId, repo.tExp).Err()
}

// TODO Check for redis.nil error
func (repo *RepositoryRedis) GetId(k string) (bool, uint64, error) {
	id, err := repo.client.Get(repo.ctx, k).Uint64()
	if err == redis.Nil {
		return false, 0, nil
	}
	if err != nil {
		repo.logger.Error("Error occured during reading id from redis", zap.String("key", k))
		return false, 0, err
	}

	return true, id, nil
}
