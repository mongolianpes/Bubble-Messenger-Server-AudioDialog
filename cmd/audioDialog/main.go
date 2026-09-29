package audiodialog

import (
	"AudioDialog/internal/service"
	"net"
)

func main() {
	service := service.NewAudioDialogService()

	go func() {
		if err := service.ExchangeAudio.StartHTTPExchangeAudioService(); err != nil {
			panic(err)
		}
	}()

	go service.Cleanup.CheckLastUsedTimeInAudioDialog()

	go func() {
		lis, err := net.Listen("tcp", ":8086")
		if err != nil {
			panic(err)
		}
		if err := service.Creator.Serve(lis); err != nil {
			panic(err)
		}
	}()
}
