package grpchandlers

import (
	"AudioDialog/internal/storage"
	"context"
	"testing"

	pb "AudioDialog/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAudioDialog(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		store := storage.NewStorage()
		h := NewGRPCHandler(store)

		resp, err := h.CreateAudioDialog(context.Background(), &pb.CreateAudioDialogRequest{})
		require.NoError(t, err)
		require.NotNil(t, resp)

		assert.NotEmpty(t, resp.DialogId)
		assert.NotEmpty(t, resp.SenderId)
		assert.NotEmpty(t, resp.ReceiverId)
		assert.NotEqual(t, resp.SenderId, resp.ReceiverId)

		store.Mu.RLock()
		_, ok := store.Dialogs[resp.DialogId]
		store.Mu.RUnlock()
		assert.True(t, ok)
	})
}
