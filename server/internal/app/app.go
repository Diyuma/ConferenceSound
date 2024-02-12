package app

import (
	"errors"
	"fmt"
	"homework/server/internal/sound"
	"time"

	"go.uber.org/zap"
)

type App struct {
	repo        sound.Repository
	sound       sound.Sound // to have ability to create sound instanses
	lastSoundId map[uint32]uint64
	logger      *zap.Logger
}

var ErrorNoSuchUserIdFound error = errors.New("can't get last sound id to that user because no such user id found")
var ErrorIncorrectSoundDuration error = errors.New("incorrect sound duration - not dividable by sound grain duration")
var ErrorExpectedSoundGrainDuration error = errors.New("expected sound duration equal to sound grain duration")

const SoundGrainDuration = 256 // in ms

func NewApp(repo sound.Repository, sound sound.Sound, logger *zap.Logger) App {
	return App{repo: repo, sound: sound, lastSoundId: make(map[uint32]uint64), logger: logger}
}

// removed cause I want to have ability to change sound interface implementations
/*func (a *App) NewSound(data *[]float32, bitRate int, duration int, authors []uint32, timeSend []uint64, logger *zap.Logger) sound.Sound {
	return a.sound.NewSound(data, bitRate, duration, authors, timeSend, logger)
}*/

func (a *App) getSoundBySoundId(soundId uint64, userId uint32, confId uint64) (*sound.Sound, uint64, error) {
	s, err := a.repo.GetSound(fmt.Sprintf("%d:%d", soundId, confId))
	a.logger.Info("getSoundBySoundId", zap.Any("sound", s))
	if err != nil {
		return nil, 0, err
	}

	if ok, timeSend := (*s).AmIAuthor(userId); ok {
		return nil, timeSend, nil
	}

	return s, 0, nil
}

func (a *App) GetSoundAvaliableTicker() *time.Ticker { // TODO change time to safe zone and non-safe and work in that way
	return time.NewTicker(time.Millisecond * SoundGrainDuration)
}

func (a *App) getLastSoundId(userId uint32) (uint64, error) { // it's bad to use locally (may be) - need to care after retrying connections
	id, ok := a.lastSoundId[userId]
	if !ok {
		return 0, ErrorNoSuchUserIdFound
	}

	return id, nil
}

func genNextAvaliableSoundId(now uint64) uint64 { // divide every 100 ms
	return (now + 99) % SoundGrainDuration
}

// ! SoundGrainDuration % genSoundDuration(...) == 0
func genSoundDuration(userId uint32, confId uint64) int { // TODO add logic there, but I can't really understand what for I may need changable sound duration
	return SoundGrainDuration
}

func (a *App) GenSoundBitRate(userId uint32, confId uint64) int { // TODO add logic there
	return (1 << 13)
}

func (a *App) GetNextSoundByUserId(userId uint32, confId uint64) (*sound.Sound, []uint64, error) { // TODO check what is better to use
	lastSoundId, err := a.getLastSoundId(userId)
	if err != nil {
		return nil, nil, err
	}

	duration := genSoundDuration(userId, confId)
	s := a.sound.NewEmptySound()
	timesSend := make([]uint64, duration/SoundGrainDuration)

	for i := 0; i < duration/SoundGrainDuration; i++ {
		nextSoundId := genNextAvaliableSoundId(lastSoundId)
		sNow, timeSend, err := a.getSoundBySoundId(nextSoundId, userId, confId)
		if err != nil {
			return nil, []uint64{}, err
		}

		lastSoundId = nextSoundId
		timesSend[i] = timeSend
		s.Append(sNow)
	}

	return &s, timesSend, nil
}

func (a *App) GetNextSoundGrainByUserId(userId uint32, confId uint64) (*sound.Sound, uint64, error) {
	lastSoundId, err := a.getLastSoundId(userId)
	if err != nil {
		return nil, 0, err
	}
	return a.getSoundBySoundId(genNextAvaliableSoundId(lastSoundId), userId, confId)
}

// It may be slow to work with error during not getting sound - may be ok will be better
func (a *App) SetSound(s sound.Sound, userId uint32, confId uint64) ([]uint64, error) { // deprecated
	if s.GetSoundDuration()%SoundGrainDuration != 0 {
		return nil, ErrorIncorrectSoundDuration
	}

	soundId := genNextAvaliableSoundId(uint64(time.Now().UnixMilli()))
	soundIds := make([]uint64, s.GetSoundDuration()/SoundGrainDuration)

	sounds, err := a.sound.DivideIntoParts(SoundGrainDuration)
	if err != nil {
		return nil, err
	}

	for i, sToAdd := range sounds {
		soundIds[i] = soundId
		sToAdd.SetTimeId(soundId)
		key := fmt.Sprintf("%d:%d", soundId, confId)
		_, sNow, err := a.repo.GetDelSound(key)
		if err != nil {
			return nil, err
		}

		(*sNow).Add(&sToAdd)

		if err = a.repo.SetSound(key, sNow, genTimeExpiration()); err != nil {
			return nil, err
		}

		soundId += SoundGrainDuration
	}

	return soundIds, nil
}

// It may be slow to work with error during not getting sound - may be ok will be better
// what will happen if 2 goroutines will run this func in same time (to sound arr)
func (a *App) SetGrainSound(s sound.Sound, userId uint32, confId uint64) (uint64, error) {
	if s == nil {
		a.logger.Error("sound is nil")
	}
	a.logger.Info("SetGrainSound", zap.Int("duration", s.GetSoundDuration()), zap.Int("bitRate", s.GetBitRate()), zap.Int("dataLen", int(len(*s.GetData()))))

	if s.GetSoundDuration() != SoundGrainDuration {
		return 0, ErrorExpectedSoundGrainDuration
	}

	soundId := genNextAvaliableSoundId(uint64(time.Now().UnixMilli()))

	s.SetTimeId(soundId)

	key := fmt.Sprintf("%d:%d", soundId, confId)
	ok, sNow, err := a.repo.GetDelSound(key)
	if err != nil {
		return 0, err
	}

	if ok {
		if sNow == nil || *sNow == nil { // there was a mistake with null pointers
			a.logger.Error("sNow is nil, but mustnt be", zap.String("Redis key", key))
			return 0, errors.New("unexpected error - sNow is nil")
		}
		a.logger.Info("sNow", zap.Any("sNow", sNow)) // !!!
		a.logger.Info("s", zap.Any("s", s))
		err = (*sNow).Add(&s)
		if err != nil {
			return 0, err
		}
	} else {
		sNow = &s
	}
	sNow = &s

	if sNow == nil {
		a.logger.Error("sNow is nil, but mustnt be", zap.String("Redis key", key))
		return 0, errors.New("unexpected error - sNow is nil")
	}

	if err := a.repo.SetSound(key, sNow, genTimeExpiration()); err != nil {
		return 0, err
	}

	return soundId, nil
}

func genTimeExpiration() time.Duration {
	return time.Second * 5
}
