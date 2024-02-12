package grpcserver

import (
	"context"
	"homework/server/internal/ports/grpcserver/proto"
	"homework/server/internal/sound/soundwav"
	"math/rand"
	"time"
)

// we choose sound impelemtation in that file

func (s *server) SendSoundDataStream(ctx context.Context, cancel context.CancelCauseFunc, userId uint32, confId uint64) <-chan *proto.ChatServerMessage {
	ticker := s.app.GetSoundAvaliableTicker()
	defer ticker.Stop()

	stream := make(chan *proto.ChatServerMessage, 10)
	go func(ctx context.Context, ticker *time.Ticker) {
		for {
			select {
			case <-ticker.C:
				sound, sId, err := s.app.GetNextSoundGrainByUserId(userId, confId)
				if err != nil {
					cancel(err)
					return
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

func (s *server) AddSoundData(data *proto.ChatClientMessage) (*proto.ClientResponseMessage, error) {
	if data == nil {
		s.logger.Error("data is nil")
	}
	if data.Data == nil {
		s.logger.Error("data.Data is nil")
	}

	userId, confId := data.UserId, data.ConfId
	sId, err := s.app.SetGrainSound(soundwav.NewSound(&data.Data, int(data.Rate), len(data.Data)/int(data.Rate), []uint32{userId}, []uint64{}), userId, confId)
	//sId, err := s.app.SetSound(soundwav.NewSound(&data.Data, int(data.Rate), len(data.Data)/int(data.Rate), []uint32{userId}, []uint64{}, s.logger), userId, confId)
	if err != nil {
		return &proto.ClientResponseMessage{Rate: 0, SoundId: 0}, err
	}

	return &proto.ClientResponseMessage{Rate: int64(s.app.GenSoundBitRate(userId, confId)), SoundId: sId}, nil
}
