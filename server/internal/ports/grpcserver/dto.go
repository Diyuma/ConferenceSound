package grpcserver

import (
	"conference/internal/app"
	"conference/internal/ports/grpcserver/proto"
	"conference/internal/sound/soundwav"
	"context"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

// we choose sound impelemtation in that file

func (s *server) SendSoundDataStream(ctx context.Context, cancel context.CancelCauseFunc, userId uint32, confId uint64) <-chan *proto.ChatServerMessage {
	ticker := s.app.GetSoundAvaliableTicker() // TODO add ability to stop ticker if service done it's work

	stream := make(chan *proto.ChatServerMessage, 10)
	go func(ctx context.Context, ticker *time.Ticker) {
		var lastSoundId uint64 = 0
		for {
			select {
			case <-ticker.C:
				sound, sId, err := s.app.GetNextSoundGrainByUserId(userId, confId, &lastSoundId)
				if err == app.ErrorNextSoundIsNotReadyYet {
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

				stream <- &proto.ChatServerMessage{Data: *(*sound).GetData(), Rate: int64((*sound).GetBitRate()), SoundId: sId}
			case <-ctx.Done():
				return
			}
		}
	}(ctx, ticker)

	return stream
}

func (s *server) GenUserId() uint32 {
	return rand.Uint32()
}

func (s *server) GenConfId() uint64 {
	return rand.Uint64()
}

// проблема в том, что клиент отправляет много в ряд и сервер не успевает их расспеределить!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
func (s *server) AddSoundData(data *proto.ChatClientMessage) (*proto.ClientResponseMessage, error) {
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
	sId, err := s.app.SetGrainSound(soundwav.NewSound(&data.Data, int(data.Rate), len(data.Data)/int(data.Rate), []uint32{userId}, []uint64{}), userId, confId, tS, mId)
	//sId, err := s.app.SetSound(soundwav.NewSound(&data.Data, int(data.Rate), len(data.Data)/int(data.Rate), []uint32{userId}, []uint64{}, s.logger), userId, confId)
	if err != nil {
		return &proto.ClientResponseMessage{Rate: 0, SoundId: 0}, err
	}

	//s.uInf.SetId(fmt.Sprint(userId, ':', confId), lastSId)

	return &proto.ClientResponseMessage{Rate: int64(s.app.GenSoundBitRate(userId, confId)), SoundId: sId}, nil
}
