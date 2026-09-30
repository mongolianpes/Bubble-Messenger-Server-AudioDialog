package audiodialog

import (
	"AudioDialog/internal/service"
	"log/slog"
	"net"
)

func main() {
	service := service.NewAudioDialogService()

	go func() {
		if err := service.ExchangeAudio.StartHTTPExchangeAudioService(); err != nil {
			slog.Error("Error start HTTP exchanger service", "error", err)
		}
	}()

	go service.Cleanup.CheckLastUsedTimeInAudioDialog()

	go func() {
		lis, err := net.Listen("tcp", ":8086")
		if err != nil {
			slog.Error("Error setup net listener for gRPC service", "error", err)
			return
		}
		if err := service.MakeDialog.Serve(lis); err != nil {
			slog.Error("Error serve gRPC service", "error", err)
		}
	}()
}
