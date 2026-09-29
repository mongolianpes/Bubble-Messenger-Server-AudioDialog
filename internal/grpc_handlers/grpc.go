package grpchandlers

import (
	"AudioDialog/internal/storage"
	"context"

	pb "AudioDialog/proto"
)

type GRPCCreatorService struct {
	pb.UnimplementedAudioDialogServiceServer
	storage storage.Creator
}

func NewGRPCHandler(storage storage.Creator) *GRPCCreatorService {
	return &GRPCCreatorService{
		storage: storage,
	}
}

func (s *GRPCCreatorService) CreateAudioDialog(ctx context.Context, req *pb.CreateAudioDialogRequest) (*pb.CreateAudioDialogResponse, error) {
	dialogID, senderID, receiverID := s.storage.CreateDialog()

	res := &pb.CreateAudioDialogResponse{
		DialogId:   dialogID,
		SenderId:   senderID,
		ReceiverId: receiverID,
	}
	return res, nil
}
