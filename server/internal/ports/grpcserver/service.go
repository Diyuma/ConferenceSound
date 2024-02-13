package grpcserver

import (
	"context"

	"conference/server/internal/ports/grpcserver/proto"

	"go.uber.org/zap"
)

func (s *server) GetSound(in *proto.ClientInfoMessage, stream proto.SoundService_GetSoundServer) error {
	//uId := s.GenUserId() // no need to gen there - it is bug
	uId := uint32(in.UserId)
	cId := in.ConfId
	ctx, cancel := context.WithCancelCause(context.Background())
	soundStream := s.SendSoundDataStream(ctx, cancel, uId, cId)
	//_ = s.SendSoundDataStream(ctx, cancel, uId, cId)?????????????????????? why i need it
	for {
		select {
		case m := <-soundStream:
			s.logger.Info("Send sound data to client: ",
				zap.Int64("rate", m.GetRate()),
				zap.Uint64("soundId", m.GetSoundId()))

			if err := stream.Send(m); err != nil {
				s.logger.Error("Failed to send a message: ", zap.Error(err))
				return err
			}
		case <-ctx.Done():
			s.logger.Warn("Ctx to send data cancelled: ", zap.Error(ctx.Err()))
			return nil

			//ctx, cancel = context.WithCancelCause(context.Background())
			//soundStream = s.SendSoundDataStream(ctx, cancel, uId, cId)?????????????????????? why i need it
		}
	}

	return nil
}

func (s *server) SendSound(ctx context.Context, in *proto.ChatClientMessage) (out *proto.ClientResponseMessage, err error) {
	defer func() {
		s.logger.Info("Send response to sound data:",
			zap.Int64("rate", out.GetRate()),
			zap.Uint64("soundId", out.GetSoundId()))
	}()

	s.logger.Info("Got sound data:",
		zap.Int64("rate", in.GetRate()),
		zap.Uint32("userId", in.GetUserId()),
		zap.Uint64("confId", in.GetConfId()))

	if s == nil {
		s.logger.Error("s is nil")
	}
	if in == nil {
		s.logger.Error("input is nil")
	}

	r, err := s.AddSoundData(in)
	if err != nil {
		s.logger.Error("Got sound data error:",
			zap.Any("response", *r),
			zap.Error(err),
		)
	}
	if r == nil {
		s.logger.Error("response is nil")
	}
	return r, err
}

func (s *server) InitUser(ctx context.Context, in *proto.EmptyMessage) (out *proto.ClientUserInitResponseMessage, err error) {
	defer func() {
		s.logger.Info("Init client send:",
			zap.Uint32("userId", out.GetUserId()))
	}()

	return &proto.ClientUserInitResponseMessage{UserId: s.GenUserId()}, nil
}

func (s *server) InitConf(ctx context.Context, in *proto.EmptyMessage) (out *proto.ClientConfInitResponseMessage, err error) {
	defer func() {
		s.logger.Info("Init conf send:",
			zap.Uint64("confId", out.GetConfId()))
	}()

	return &proto.ClientConfInitResponseMessage{ConfId: s.GenConfId()}, nil
}
