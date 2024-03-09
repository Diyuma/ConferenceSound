package app

import (
	"conference/internal/sound"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type App struct {
	repo   sound.Repository
	sound  sound.Sound // to have ability to create sound instanses
	logger *zap.Logger
}

var ErrorNoSuchUserIdFound error = errors.New("can't get last sound id to that user because no such user id found")
var ErrorNoSuchSoundIdFound error = errors.New("can't get sound because no such sound id found")
var ErrorIncorrectSoundDuration error = errors.New("incorrect sound duration - not dividable by sound grain duration")
var ErrorExpectedSoundGrainDuration error = errors.New("expected sound duration equal to sound grain duration")
var ErrorExpectedNonNillObject error = errors.New("expected to work with real object, not nill")
var ErrorNextSoundIsNotReadyYet error = errors.New("requested sound id to send may be counting on server right now")

const SoundGrainDuration = 256 // in ms
const MinimumBitRate = 4096
const TimeToSaveSound = 2047 // in ms

func NewApp(repo sound.Repository, sound sound.Sound, logger *zap.Logger) App {
	return App{repo: repo, sound: sound, logger: logger}
}

func (a *App) getSoundBySoundId(soundId uint64, userId uint32, confId uint64) (*sound.Sound, uint64, error) {
	ok, s, err := a.repo.GetSound(fmt.Sprintf("%d:%d", soundId, confId))
	if !ok && err == nil {
		return nil, 0, ErrorNoSuchSoundIdFound
	}
	a.logger.Info("getSoundBySoundId", zap.Int("soundDuration", (*s).GetSoundDuration()), zap.Int("bitrate", (*s).GetBitRate()), zap.Any("sound", (*s).GetAuthors()), zap.Any("timeIds", (*s).GetTimeId()))
	if err != nil {
		return nil, 0, err
	}

	if ok, timeSend := (*s).AmIAuthor(userId); ok { // TODO I think that it is correct to send s , not nill there - check it
		return s, timeSend, nil
	}

	return s, 0, nil
}

func (a *App) GetSoundAvaliableTicker() *time.Ticker {
	return time.NewTicker(time.Millisecond * SoundGrainDuration)
}

func genNextAvaliableSoundIdToRead(now uint64) uint64 { // TODO change type of sound id to smth less ; it is dagnerous idea because what if server get 2 in a row form 1 user
	return now / SoundGrainDuration
}

func genNextAvaliableSoundIdToWrite(now uint64, mId uint32) uint64 { // TODO change type of sound id to smth less ; it is dagnerous idea because what if server get 2 in a row form 1 user
	return now/SoundGrainDuration + 2 + uint64(mId)
}

func isSoundIdTooOldToRead(now uint64, sId uint64) (uint64, bool) {
	lSId := genNextAvaliableSoundIdToRead(now)
	if lSId > sId+TimeToSaveSound/SoundGrainDuration {
		return lSId, true
	}
	return 0, false
}

func isSoundIdTooOldToWrite(now uint64, sId uint64, mId uint32) (uint64, bool) {
	lSId := genNextAvaliableSoundIdToRead(now)
	if lSId+1 > sId { // TODO why mId not adding there???
		return lSId + uint64(mId), true
	}
	return 0, false
}

// ! SoundGrainDuration % genSoundDuration(...) == 0
func genSoundDuration(userId uint32, confId uint64) int { // TODO add logic there, but I can't really understand what for I may need changable sound duration
	return SoundGrainDuration
}

func (a *App) GenSoundBitRate(userId uint32, confId uint64, now uint64, tS uint64, mId uint64, cBr int) int { // TODO add logic there
	return 4096 * 2 * 2
	//return cBr
	diff := now - tS
	if diff < SoundGrainDuration/3 {
		cBr *= 2
	}
	for diff >= SoundGrainDuration/2 && cBr > MinimumBitRate {
		cBr /= 2
		diff -= SoundGrainDuration / 2
	}
	return cBr
}

func (a *App) GetNextSoundGrainByUserId(uId uint32, cId uint64, lSId *uint64) (*sound.Sound, uint64, error) {
	a.logger.Info("GetNextGrainSound", zap.Uint32("userId", uId), zap.Uint64("confId", cId))
	if id, ok := isSoundIdTooOldToRead(uint64(time.Now().UnixMilli()), *lSId); ok {
		(*lSId) = id
	}
	defer func() { (*lSId)++ }()
	/*lastSoundId, err := a.getLastSoundId(userId)
	if err != nil {
		return nil, 0, err
	}*/
	s, sId, err := a.getSoundBySoundId(*lSId, uId, cId)
	if err == ErrorNoSuchSoundIdFound {
		a.logger.Warn("No such sound id found", zap.Uint32("userId", uId), zap.Uint64("confId", cId), zap.Uint64("soundId", *lSId))
		return nil, 0, ErrorNextSoundIsNotReadyYet
	}
	return s, sId, err
}

// It may be slow to work with error during not getting sound - may be ok will be better
// what will happen if 2 goroutines will run this func in same time (to sound arr)
func (a *App) SetGrainSound(s sound.Sound, userId uint32, confId uint64, tS uint64, mId uint32, tNow uint64) (uint64, error) {
	if s == nil {
		a.logger.Warn(ErrorExpectedNonNillObject.Error(), zap.Uint32("userId", userId), zap.Uint64("confId", confId))
		return 0, ErrorExpectedNonNillObject
	}

	a.logger.Info("SetGrainSound", zap.Int("duration", s.GetSoundDuration()), zap.Int("bitRate", s.GetBitRate()), zap.Int("dataLen", int(len(*s.GetData()))),
		zap.Uint32("userId", userId), zap.Uint64("confId", confId), zap.Uint64("timestamp", tS), zap.Uint32("messageId", mId))

	if s.GetSoundDuration() != SoundGrainDuration {
		a.logger.Warn("Unexpected sound duration")
		return 0, ErrorExpectedSoundGrainDuration
	}

	sId := genNextAvaliableSoundIdToWrite(tS, mId)
	if id, ok := isSoundIdTooOldToWrite(tNow, sId, mId); ok {
		sId = id
	}

	s.SetTimeId(sId)

	key := fmt.Sprintf("%d:%d", sId, confId)
	ok, sNow, err := a.repo.GetDelSound(key)
	if err != nil {
		a.logger.Warn("Tried to get current sound from repo, got error", zap.Error(err))
		return 0, err
	}

	if ok {
		err = (*sNow).Add(&s)
		if err != nil {
			a.logger.Warn("Tried to add one sound to another, got error", zap.Error(err))
			return 0, err
		}
	} else {
		sNow = &s
	}

	if err := a.repo.SetSound(key, sNow, genTimeExpiration()); err != nil {
		a.logger.Error("Tried to set new sound, got error", zap.Error(err))
		return 0, err
	}

	return sId, nil
}

func genTimeExpiration() time.Duration {
	return time.Second * 5
}
