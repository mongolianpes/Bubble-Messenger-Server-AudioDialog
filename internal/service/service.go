package service

import (
	grpcHandlers "AudioDialog/internal/grpc_handlers"
	httpHandlers "AudioDialog/internal/http_handlers"
	"AudioDialog/internal/storage"

	pb "AudioDialog/proto"

	"google.golang.org/grpc"
)

type AudioDialogService struct {
	MakeDialog    *grpc.Server
	ExchangeAudio httpHandlers.ExchangeAudioService
	Cleanup       storage.CleanupDialogs
}

func NewAudioDialogService() *AudioDialogService {
	storage := storage.NewStorage()
	exchangeAudio := httpHandlers.NewHTTPExchangeAudioService(storage)

	grpcServer := grpc.NewServer()
	pb.RegisterAudioDialogServiceServer(grpcServer, grpcHandlers.NewGRPCHandler(storage))

	return &AudioDialogService{
		MakeDialog:    grpcServer,
		ExchangeAudio: exchangeAudio,
		Cleanup:       storage,
	}
}
