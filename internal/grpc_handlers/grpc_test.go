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

func TestDeleteAudioDialog(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		store := storage.NewStorage()
		h := NewGRPCHandler(store)

		created, err := h.CreateAudioDialog(context.Background(), &pb.CreateAudioDialogRequest{})
		require.NoError(t, err)

		resp, err := h.DeleteAudioDialog(context.Background(), &pb.DeleteAudioDialogRequest{
			DialogId: created.DialogId,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)

		store.Mu.RLock()
		_, ok := store.Dialogs[created.DialogId]
		store.Mu.RUnlock()
		assert.False(t, ok)
	})

	t.Run("idempotent — missing dialog", func(t *testing.T) {
		store := storage.NewStorage()
		h := NewGRPCHandler(store)

		resp, err := h.DeleteAudioDialog(context.Background(), &pb.DeleteAudioDialogRequest{
			DialogId: "no-such",
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("empty dialog id", func(t *testing.T) {
		store := storage.NewStorage()
		h := NewGRPCHandler(store)
		_, senderID, _ := store.CreateDialog() // диалог в store есть, но id другой

		resp, err := h.DeleteAudioDialog(context.Background(), &pb.DeleteAudioDialogRequest{
			DialogId: "",
		})
		require.NoError(t, err)
		require.NotNil(t, resp)

		// существующие диалоги не затронуты
		store.Mu.RLock()
		assert.NotEmpty(t, store.Dialogs)
		store.Mu.RUnlock()
		_ = senderID
	})
}
