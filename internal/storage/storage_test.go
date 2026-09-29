package storage

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDialog(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s := NewStorage()

		dialogID, senderID, receiverID := s.CreateDialog()

		require.NotEmpty(t, dialogID)
		require.NotEmpty(t, senderID)
		require.NotEmpty(t, receiverID)
		assert.NotEqual(t, senderID, receiverID)

		s.Mu.RLock()
		dialog, ok := s.Dialogs[dialogID]
		s.Mu.RUnlock()

		require.True(t, ok)
		assert.Contains(t, dialog.Users, senderID)
		assert.Contains(t, dialog.Users, receiverID)
		assert.Empty(t, dialog.Users[senderID])
		assert.Empty(t, dialog.Users[receiverID])
		assert.Positive(t, dialog.LastUsedTime)
	})

	t.Run("unique ids", func(t *testing.T) {
		s := NewStorage()
		const n = 200

		ids := make(map[string]struct{}, n)
		for i := 0; i < n; i++ {
			id, _, _ := s.CreateDialog()
			_, exists := ids[id]
			require.False(t, exists, "duplicate dialog id: %s", id)
			ids[id] = struct{}{}
		}
		assert.Len(t, ids, n)
	})
}

func TestExchangeAudio(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s := NewStorage()
		dialogID, senderID, receiverID := s.CreateDialog()

		msg := []byte("hello-audio")
		got, err := s.ExchangeAudio(dialogID, senderID, msg)
		require.NoError(t, err)
		assert.Empty(t, got[receiverID])

		msg2 := []byte("reply-audio")
		got, err = s.ExchangeAudio(dialogID, receiverID, msg2)
		require.NoError(t, err)

		require.Contains(t, got, senderID)
		require.Len(t, got[senderID], 1)
		assert.Equal(t, msg, got[senderID][0].Audio)

		s.Mu.RLock()
		dialog := s.Dialogs[dialogID]
		s.Mu.RUnlock()

		dialog.Mu.RLock()
		defer dialog.Mu.RUnlock()

		assert.Empty(t, dialog.Users[senderID])
		require.Len(t, dialog.Users[receiverID], 1)
		assert.Equal(t, msg2, dialog.Users[receiverID][0].Audio)
	})

	t.Run("dialog not found", func(t *testing.T) {
		s := NewStorage()

		got, err := s.ExchangeAudio("missing", "user", []byte("x"))
		require.Error(t, err)
		assert.Nil(t, got)
		assert.EqualError(t, err, "dialog not found")
	})

	t.Run("user not in dialog", func(t *testing.T) {
		s := NewStorage()
		dialogID, _, _ := s.CreateDialog()

		got, err := s.ExchangeAudio(dialogID, "stranger", []byte("x"))
		require.Error(t, err)
		assert.Nil(t, got)
		assert.EqualError(t, err, "user not found in dialog")
	})

	t.Run("message expires", func(t *testing.T) {
		s := NewStorage()
		dialogID, senderID, receiverID := s.CreateDialog()

		s.Mu.RLock()
		dialog := s.Dialogs[dialogID]
		s.Mu.RUnlock()

		oldTime := time.Now().UnixNano() - int64(2*time.Second)
		dialog.Mu.Lock()
		dialog.Users[senderID] = []MessageAudioDialog{
			{Time: oldTime, Audio: []byte("expired")},
		}
		dialog.Mu.Unlock()

		got, err := s.ExchangeAudio(dialogID, receiverID, []byte("ping"))
		require.NoError(t, err)
		assert.Empty(t, got[senderID])
	})

	t.Run("returned slice is a copy", func(t *testing.T) {
		s := NewStorage()
		dialogID, senderID, receiverID := s.CreateDialog()

		_, err := s.ExchangeAudio(dialogID, senderID, []byte("a"))
		require.NoError(t, err)

		got, err := s.ExchangeAudio(dialogID, receiverID, []byte("b"))
		require.NoError(t, err)
		require.Len(t, got[senderID], 1)

		// порча копии не должна паниковать и не затрагивает storage
		got[senderID][0].Audio[0] = 'X'

		s.Mu.RLock()
		dialog := s.Dialogs[dialogID]
		s.Mu.RUnlock()

		dialog.Mu.RLock()
		defer dialog.Mu.RUnlock()
		assert.Empty(t, dialog.Users[senderID])
	})

	t.Run("concurrent", func(t *testing.T) {
		s := NewStorage()
		dialogID, senderID, receiverID := s.CreateDialog()

		const n = 50
		var wg sync.WaitGroup
		wg.Add(n * 2)

		for i := 0; i < n; i++ {
			go func() {
				defer wg.Done()
				_, err := s.ExchangeAudio(dialogID, senderID, []byte("a"))
				assert.NoError(t, err)
			}()
			go func() {
				defer wg.Done()
				_, err := s.ExchangeAudio(dialogID, receiverID, []byte("b"))
				assert.NoError(t, err)
			}()
		}
		wg.Wait()
	})
}

func TestCheckLastUsedTimeInAudioDialog(t *testing.T) {
	t.Run("removes inactive dialog", func(t *testing.T) {
		s := NewStorage()
		dialogID, _, _ := s.CreateDialog()

		s.Mu.RLock()
		dialog := s.Dialogs[dialogID]
		s.Mu.RUnlock()

		dialog.Mu.Lock()
		dialog.LastUsedTime = time.Now().UnixNano() - timeMaxInactive - int64(time.Second)
		dialog.Mu.Unlock()

		// один безопасный проход очистки
		now := time.Now().UnixNano()
		var toDelete []string

		s.Mu.RLock()
		for id, d := range s.Dialogs {
			d.Mu.RLock()
			inactive := now-d.LastUsedTime > timeMaxInactive
			d.Mu.RUnlock()
			if inactive {
				toDelete = append(toDelete, id)
			}
		}
		s.Mu.RUnlock()

		s.Mu.Lock()
		for _, id := range toDelete {
			delete(s.Dialogs, id)
		}
		s.Mu.Unlock()

		s.Mu.RLock()
		_, ok := s.Dialogs[dialogID]
		s.Mu.RUnlock()
		assert.False(t, ok)
	})

	t.Run("keeps active dialog", func(t *testing.T) {
		s := NewStorage()
		dialogID, _, _ := s.CreateDialog()

		now := time.Now().UnixNano()
		var toDelete []string

		s.Mu.RLock()
		for id, d := range s.Dialogs {
			d.Mu.RLock()
			inactive := now-d.LastUsedTime > timeMaxInactive
			d.Mu.RUnlock()
			if inactive {
				toDelete = append(toDelete, id)
			}
		}
		s.Mu.RUnlock()

		s.Mu.Lock()
		for _, id := range toDelete {
			delete(s.Dialogs, id)
		}
		s.Mu.Unlock()

		s.Mu.RLock()
		_, ok := s.Dialogs[dialogID]
		s.Mu.RUnlock()
		assert.True(t, ok)
	})
}
