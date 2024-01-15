package grpcserver

import (
	"context"

	"homework/server/internal/ports/grpcserver/proto"

	"go.uber.org/zap"
)

func (s *server) GetSound(in *proto.ClientInfoMessage, stream proto.SoundService_GetSoundServer) error {
	uId := s.GenUserId()
	cId := in.ConfId
	ctx, cancel := context.WithCancelCause(context.Background())
	soundStream := s.SendSoundDataStream(ctx, cancel, uId, cId)
	for {
		select {
		case m := <-soundStream:
			if err := stream.Send(m); err != nil {
				s.logger.Warn("Failed to send a message: ", zap.Error(err))
				return err
			}
		case <-ctx.Done():
			s.logger.Warn("Ctx to send data cancelled: ", zap.Error(ctx.Err()))

			ctx, cancel = context.WithCancelCause(context.Background())
			soundStream = s.SendSoundDataStream(ctx, cancel, uId, cId)
		}
	}

	return nil
}

func (s *server) SendSound(ctx context.Context, in *proto.ChatClientMessage) (*proto.ClientResponseMessage, error) {
	return s.AddSoundData(in)
}

func (s *server) InitClient(ctx context.Context, in *proto.ClientInfoMessage) (*proto.ClientInitResponseMessage, error) {
	return &proto.ClientInitResponseMessage{ClientId: s.GenUserId(), ConfId: s.GenConfId()}, nil
}
