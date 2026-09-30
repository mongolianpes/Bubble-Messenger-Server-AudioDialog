package grpchandlers

import (
	"AudioDialog/internal/storage"
	"context"
	"log/slog"

	pb "AudioDialog/proto"
)

type GRPCMakeDialogService struct {
	pb.UnimplementedAudioDialogServiceServer
	storage storage.MakeDialog
}

func NewGRPCHandler(storage storage.MakeDialog) *GRPCMakeDialogService {
	return &GRPCMakeDialogService{
		storage: storage,
	}
}

func (s *GRPCMakeDialogService) CreateAudioDialog(ctx context.Context, req *pb.CreateAudioDialogRequest) (*pb.CreateAudioDialogResponse, error) {
	dialogID, senderID, receiverID := s.storage.CreateDialog()

	slog.Info("CreateAudioDialog", "dialogID", dialogID, "user1", senderID, "user2", receiverID)

	res := &pb.CreateAudioDialogResponse{
		DialogId:   dialogID,
		SenderId:   senderID,
		ReceiverId: receiverID,
	}
	return res, nil
}

func (s *GRPCMakeDialogService) DeleteAudioDialog(ctx context.Context, req *pb.DeleteAudioDialogRequest) (*pb.DeleteAudioDialogResponse, error) {
	s.storage.DeleteDialog(req.DialogId)

	slog.Info("DeleteAudioDialog", "dialogID", req.DialogId)

	return &pb.DeleteAudioDialogResponse{}, nil
}
