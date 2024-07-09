package grpcserver

import (
	"conference/internal/app"
	"conference/internal/ports/grpcserver/protosound"
	"conference/internal/sound/soundwav"
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

// we choose sound impelemtation in that file

func (s *Server) SendSoundDataStream(ctx context.Context, cancel context.CancelCauseFunc, userId uint32, confId uint64) <-chan *protosound.ChatServerMessage {
	ticker := s.app.GetSoundAvaliableTicker() // TODO add ability to stop ticker if service done it's work

	stream := make(chan *protosound.ChatServerMessage, 10)
	go func(ctx context.Context, ticker *time.Ticker) {
		var lastSoundId uint64 = 0
		for {
			select {
			case <-ticker.C:
				ok, br, err := s.uInf.GetBitRate(fmt.Sprint(userId, ':', confId))
				if !ok || err != nil {
					br = -1
				}
				br = -1 // care it is timed solution !!!
				sound, sId, onlyOne, err := s.app.GetNextSoundGrainByUserId(userId, confId, &lastSoundId, br)
				if err == app.ErrorNextSoundIsNotReadyYet {
					s.logger.Debug("next sound data is not ready yet")
					continue
				}
				if err != nil {
					s.logger.Error(err.Error())
					cancel(err)
					return
				}

				if sound == nil || *sound == nil {
					s.logger.Warn("Tried to send nil sound data to client", zap.Uint32("userId", userId), zap.Uint64("confId", confId))
				}

				s.logger.Info("Send sound to client", zap.Uint32("userId", userId), zap.Uint64("confId", confId), zap.Int64("rate", int64((*sound).GetBitRate())), zap.Any("soundIds", (*sound).GetTimeId()))

				stream <- &protosound.ChatServerMessage{Data: *(*sound).GetData(), Rate: int64((*sound).GetBitRate()), SoundId: sId, OnlyOne: onlyOne}
			case <-ctx.Done():
				return
			}
		}
	}(ctx, ticker)

	return stream
}

func (s *Server) GenUserId() uint32 {
	return rand.Uint32()
}

func (s *Server) GenConfId() uint64 {
	return rand.Uint64()
}

func (s *Server) AddSoundData(data *protosound.ChatClientMessage) (*protosound.ClientResponseMessage, error) {
	if data == nil {
		s.logger.Error("data is nil")
		return nil, app.ErrorExpectedNonNillObject
	}
	if data.Data == nil {
		s.logger.Error("data.Data is nil")
	}

	userId, confId := data.UserId, data.ConfId
	/*ok, lastSId, err := s.uInf.GetId(fmt.Sprint(userId, ':', confId)) // TODO READ FROM DB
	if !ok && err == nil {
		lastSId = 0
	}
	if err != nil {
		return nil, err
	}*/
	tS, mId := data.TimeStamp, data.MessageInd
	tNow := uint64(time.Now().UnixMilli())
	sId, err := s.app.SetGrainSound(soundwav.NewSound(&data.Data, int(data.Rate), len(data.Data)/int(data.Rate), []uint32{userId}, []uint64{}), userId, confId, tS, mId, tNow)
	if err != nil {
		s.logger.Warn("failed to set grain sound", zap.Error(err))
		return &protosound.ClientResponseMessage{Rate: 0, SoundId: 0, MessageInd: 0}, err
	}

	ok, preferredBr, err := s.uInf.GetBitRate(fmt.Sprint(userId, ':', confId))
	if ok && err == nil {
		if data.Rate < int64(preferredBr) {
			s.ChangeUserBitRate(userId, confId, min(app.MaximumBitRate, max(app.MinimumBitRate, int(data.Rate))))
		}
		if data.Rate > int64(preferredBr) {
			s.ChangeUserBitRate(userId, confId, min(app.MaximumBitRate, max(app.MinimumBitRate, preferredBr*2)))
		}
	}

	//s.uInf.SetId(fmt.Sprint(userId, ':', confId), lastSId)

	return &protosound.ClientResponseMessage{Rate: int64(s.app.GenSoundBitRate(userId, confId, tNow, tS, uint64(mId), int(data.Rate))), SoundId: sId, MessageInd: data.GetMessageInd()}, nil
}

func (s *Server) ChangeUserBitRate(uId uint32, cId uint64, br int) error { // br = 0 is ok cause we take maximum
	s.logger.Debug("change user bitrate", zap.Uint32("uId", uId), zap.Uint64("cId", cId), zap.Int("bitrate", br))
	if err := s.uInf.SetBitRate(fmt.Sprint(uId, ':', cId), min(app.MaximumBitRate, max(app.MinimumBitRate, br))); err != nil {
		s.logger.Warn("Failed to set br to uInf repo", zap.Uint32("userId", uId), zap.Uint64("confId", cId), zap.Error(err))
	}
	return nil
}
