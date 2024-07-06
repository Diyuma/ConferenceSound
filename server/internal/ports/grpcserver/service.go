package grpcserver

import (
	"context"

	"conference/internal/ports/grpcserver/protosound"

	"go.uber.org/zap"
)

func (s *Server) GetSound(in *protosound.ClientInfoMessage, stream protosound.SoundService_GetSoundServer) error {
	//uId := s.GenUserId() // no need to gen there - it is bug
	uId := in.UserId
	cId := in.ConfId
	ctx, cancel := context.WithCancelCause(context.Background())
	soundStream := s.SendSoundDataStream(ctx, cancel, uId, cId)
	s.logger.Info("sound stream established")
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

	s.logger.Error("GetSound failed")

	return nil
}

func (s *Server) PingServer(ctx context.Context, in *protosound.ClientInfoMessage) (out *protosound.EmptyMessage, err error) {
	go s.ChangeUserBitRate(in.GetUserId(), in.GetConfId(), 0)
	return &protosound.EmptyMessage{}, nil
}

func (s *Server) SendSound(ctx context.Context, in *protosound.ChatClientMessage) (out *protosound.ClientResponseMessage, err error) {
	defer func() {
		s.logger.Info("Send response to sound data:",
			zap.Int64("rate", out.GetRate()),
			zap.Uint64("soundId", out.GetSoundId()))
	}()

	s.logger.Info("Got sound data:",
		zap.Int64("rate", in.GetRate()),
		zap.Uint32("userId", in.GetUserId()),
		zap.Uint64("confId", in.GetConfId()),
		zap.Uint32("messageId", in.GetMessageInd()),
		zap.Uint64("timeStamp", in.GetTimeStamp()))

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

func (s *Server) InitUser(ctx context.Context, in *protosound.EmptyMessage) (out *protosound.ClientUserInitResponseMessage, err error) {
	defer func() {
		s.logger.Info("Init client send:",
			zap.Uint32("userId", out.GetUserId()))
	}()

	return &protosound.ClientUserInitResponseMessage{UserId: s.GenUserId()}, nil
}

func (s *Server) InitConf(ctx context.Context, in *protosound.EmptyMessage) (out *protosound.ClientConfInitResponseMessage, err error) {
	defer func() {
		s.logger.Info("Init conf send:",
			zap.Uint64("confId", out.GetConfId()))
	}()

	return &protosound.ClientConfInitResponseMessage{ConfId: s.GenConfId()}, nil
}
