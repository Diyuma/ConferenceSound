package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"testgrpc/proto"

	"google.golang.org/grpc"
)

type server struct {
	proto.UnimplementedSoundServiceServer
}

// just to understand how it work - you may use evans (https://github.com/ktr0731/evans) and run it like
// evans --proto proto/sound_data_streaming.proto --port 8080
// proto file is in proto folder

// how it works: you need to find how to call grpc funcs and there (in proto folder in sound_data_streaming.protoc) you can find funcs
// that you may call and structs that you may use to that funcs as params

// you will get rate 5000 and data Hello in bytes or smth like SGVsbG8= if you use secure connection (base64, but I don't know for 100%)
func (s *server) GetSoundJustHello(stream proto.SoundService_GetSoundJustHelloServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Failed to receive a message : %v", err)
			return err
		}

		log.Printf("Received byte message (converted to string): %s", string(in.GetData()))

		if err := stream.Send(&proto.ChatServerMessage{Data: []byte("Hello"), Rate: 5000}); err != nil {
			log.Fatalf("Failed to send a message: %v", err)
			return err
		}
	}
}

// you will get rate 5000 and real .wav data
// If you want to send to server and get from him that sound - just send your data in data field, with your rate, and server will return it in same format
func (s *server) GetSound(stream proto.SoundService_GetSoundServer) error {
	preparedData, err := os.ReadFile("output5000.wav")
	fmt.Println(string(preparedData[:100]))
	if err != nil {
		log.Fatalf("Failed to parse file: %v", err)
		return err
	}

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Failed to receive a message : %v", err)
			return err
		}

		log.Printf("Received byte message (converted to string): %s", string(in.GetData()))

		dataToSend := preparedData
		rateToSend := int64(5000)
		if in.GetReadyToSend() {
			dataToSend = in.GetData()
			rateToSend = in.GetRate()
		}

		if err := stream.Send(&proto.ChatServerMessage{Data: dataToSend, Rate: rateToSend}); err != nil {
			log.Fatalf("Failed to send a message: %v", err)
			return err
		}
	}
}

func main() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	proto.RegisterSoundServiceServer(s, &server{})

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
