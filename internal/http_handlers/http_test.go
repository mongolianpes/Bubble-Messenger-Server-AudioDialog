package httphandlers

import (
	"AudioDialog/internal/storage"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExchangeAudioHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		store := storage.NewStorage()
		dialogID, senderID, receiverID := store.CreateDialog()
		svc := NewHTTPExchangeAudioService(store)

		e := echo.New()
		e.POST("/exchangeaudio", svc.ExchangeAudioHandler)

		body, _ := json.Marshal(map[string]any{
			"dialog_id": dialogID,
			"user_id":   senderID,
			"message":   []byte("audio-data"),
		})

		req := httptest.NewRequest(http.MethodPost, "/exchangeaudio", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string][]storage.MessageAudioDialog
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

		// ключ receiver есть, но сообщений ещё нет
		require.Contains(t, resp, receiverID)
		assert.Empty(t, resp[receiverID])
		assert.NotContains(t, resp, senderID)

		// receiver забирает сообщение sender'а
		body2, _ := json.Marshal(map[string]any{
			"dialog_id": dialogID,
			"user_id":   receiverID,
			"message":   []byte("reply"),
		})
		req2 := httptest.NewRequest(http.MethodPost, "/exchangeaudio", bytes.NewReader(body2))
		req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec2 := httptest.NewRecorder()
		e.ServeHTTP(rec2, req2)

		assert.Equal(t, http.StatusOK, rec2.Code)

		var resp2 map[string][]storage.MessageAudioDialog
		require.NoError(t, json.Unmarshal(rec2.Body.Bytes(), &resp2))
		require.Contains(t, resp2, senderID)
		require.Len(t, resp2[senderID], 1)
		assert.Equal(t, []byte("audio-data"), resp2[senderID][0].Audio)
	})

	t.Run("invalid json", func(t *testing.T) {
		svc := NewHTTPExchangeAudioService(storage.NewStorage())

		e := echo.New()
		e.POST("/exchangeaudio", svc.ExchangeAudioHandler)

		req := httptest.NewRequest(http.MethodPost, "/exchangeaudio", bytes.NewReader([]byte("{")))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("dialog not found", func(t *testing.T) {
		svc := NewHTTPExchangeAudioService(storage.NewStorage())

		e := echo.New()
		e.POST("/exchangeaudio", svc.ExchangeAudioHandler)

		body, _ := json.Marshal(map[string]any{
			"dialog_id": "no-such",
			"user_id":   "u1",
			"message":   []byte("x"),
		})

		req := httptest.NewRequest(http.MethodPost, "/exchangeaudio", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("user not in dialog", func(t *testing.T) {
		store := storage.NewStorage()
		dialogID, _, _ := store.CreateDialog()
		svc := NewHTTPExchangeAudioService(store)

		e := echo.New()
		e.POST("/exchangeaudio", svc.ExchangeAudioHandler)

		body, _ := json.Marshal(map[string]any{
			"dialog_id": dialogID,
			"user_id":   "stranger",
			"message":   []byte("x"),
		})

		req := httptest.NewRequest(http.MethodPost, "/exchangeaudio", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
