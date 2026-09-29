package httphandlers

import (
	"AudioDialog/internal/storage"
	"net/http"

	"github.com/labstack/echo/v4"
)

type HTTPExchangeAudioService struct {
	storage storage.ExchangeAudio
}

type ExchangeAudioService interface {
	StartHTTPExchangeAudioService() error
}

type ExchangeAudioRequest struct {
	DialogID string `json:"dialog_id"`
	UserID   string `json:"user_id"`
	Message  []byte `json:"message"`
}

const httpServicePort = ":8080"

func NewHTTPExchangeAudioService(storage storage.ExchangeAudio) *HTTPExchangeAudioService {
	return &HTTPExchangeAudioService{
		storage: storage,
	}
}

func (s *HTTPExchangeAudioService) StartHTTPExchangeAudioService() error {
	echo := echo.New()
	echo.POST("/exchangeaudio", s.ExchangeAudioHandler)
	return echo.Start(httpServicePort)
}

func (s *HTTPExchangeAudioService) ExchangeAudioHandler(c echo.Context) error {
	req := ExchangeAudioRequest{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}

	resp, err := s.storage.ExchangeAudio(req.DialogID, req.UserID, req.Message)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to exchange audio"})
	}

	return c.JSON(http.StatusOK, resp)
}
