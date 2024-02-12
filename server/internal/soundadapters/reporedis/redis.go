package reporedis

import (
	"bytes"
	"context"
	"encoding/gob"
	"homework/server/internal/sound"
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
	return &RepositoryRedis{ctx: ctx, client: client, logger: lr}
}

func (repo *RepositoryRedis) SetSound(k string, v *sound.Sound, tExp time.Duration) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(*v); err != nil {
		return err
	}

	return repo.client.Set(repo.ctx, k, buf.String(), tExp).Err()
}

// TODO Check for redis.nil error
func (repo *RepositoryRedis) GetSound(k string) (*sound.Sound, error) {
	res, err := repo.client.Get(repo.ctx, k).Bytes()
	if err != nil {
		return nil, err
	}

	var buf *bytes.Buffer = bytes.NewBuffer(res)
	enc := gob.NewDecoder(buf)

	var sw sound.Sound
	if err := enc.Decode(sw); err != nil {
		return nil, err
	}

	return &sw, nil
}

func (repo *RepositoryRedis) GetDelSound(k string) (bool, *sound.Sound, error) {
	res, err := repo.client.GetDel(repo.ctx, k).Bytes()
	if err == redis.Nil {
		return false, nil, nil
	}
	if err != nil {
		return false, nil, err
	}

	var buf *bytes.Buffer = bytes.NewBuffer(res)
	enc := gob.NewDecoder(buf)

	var sw sound.Sound
	if err := enc.Decode(sw); err != nil {
		return false, nil, err
	}

	return true, &sw, nil
}

/*func (repo *RepositoryRedis) Add(k int, sw sound.Sound, tExp time.Duration) error {
	swNow, err := repo.getDelSound(k) // Am I really need to del there because of small expiration time
	if err != nil {
		return err
	}

	swNow.Add(sw)

	return repo.SetSound(k, swNow, tExp)
}*/
