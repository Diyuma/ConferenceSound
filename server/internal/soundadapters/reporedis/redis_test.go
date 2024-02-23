package reporedis

import (
	"conference/internal/sound"
	"conference/internal/sound/soundwav"
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type BasicTestSuite struct {
	suite.Suite
	r *RepositoryRedis
	t *testing.T
}

func (bts *BasicTestSuite) SetupSuite() {
	lgr, err := zap.NewProduction(zap.WithCaller(true))
	if err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}

	bts.r = NewRepo(context.Background(), ":8088", lgr)
}

func (ssts *BasicTestSuite) TearDownSuite() {

}

func (bts *BasicTestSuite) TestAddAndGetSound() {
	data := []float32{1, 2, 3}
	s := sound.Sound(soundwav.NewSound(&data, 123, 456, []uint32{789, 1011}, []uint64{1213}))
	k := "key"
	assert.NoError(bts.t, bts.r.SetSound(k, &s, time.Second))

	ok, sGet, err := bts.r.GetSound(k)
	assert.NoError(bts.t, err)
	assert.Equal(bts.t, true, ok)
	assert.NotNil(bts.t, sGet)
	if sGet != nil {
		assert.Equal(bts.t, s, *sGet)
	}
}

func (bts *BasicTestSuite) TestGetDelSound() {
	data := []float32{1, 2, 3}
	s := sound.Sound(soundwav.NewSound(&data, 123, 456, []uint32{789, 1011}, []uint64{1213}))
	k := "key"
	assert.NoError(bts.t, bts.r.SetSound(k, &s, time.Second))

	ok, sGet, err := bts.r.GetDelSound(k)

	assert.NoError(bts.t, err)
	assert.Equal(bts.t, true, ok)
	assert.NotNil(bts.t, sGet)

	if sGet != nil {
		assert.Equal(bts.t, s, *sGet)
	}

	ok, sGet, err = bts.r.GetSound(k)
	assert.Nil(bts.t, err)
	assert.Equal(bts.t, false, ok)
	assert.Nil(bts.t, sGet)
}

func TestAddAndGetSound(t *testing.T) {
	bts := BasicTestSuite{t: t}
	suite.Run(t, &bts)
}
