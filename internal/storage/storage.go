package storage

import (
	"crypto/rand"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	timeToSendMessage          = int64(time.Second)
	timeMaxInactive            = int64(15 * time.Second)
	timeSleepCheckLastUsedTime = time.Second * 20
)

type Storage struct {
	Dialogs map[string]*Dialog
	Mu      sync.RWMutex
}

type Dialog struct {
	Users        map[string][]MessageAudioDialog
	LastUsedTime int64
	Mu           sync.RWMutex
}

type MakeDialog interface {
	CreateDialog() (dialogID, senderID, receiverID string)
	DeleteDialog(dialogID string)
}

type ExchangeAudio interface {
	ExchangeAudio(idDialog, userID string, message []byte) (map[string][]MessageAudioDialog, error)
}

type CleanupDialogs interface {
	CheckLastUsedTimeInAudioDialog()
}

type MessageAudioDialog struct {
	Time  int64
	Audio []byte
}

func NewStorage() *Storage {
	return &Storage{
		Dialogs: make(map[string]*Dialog),
	}
}

func (s *Storage) CreateDialog() (dialogID, senderID, receiverID string) {
	for {
		dialogID = uuid.New().String()

		s.Mu.RLock()
		_, exists := s.Dialogs[dialogID]
		s.Mu.RUnlock()

		if !exists {
			break
		}
	}

	senderID = rand.Text()
	receiverID = rand.Text()

	dialog := &Dialog{
		Users: map[string][]MessageAudioDialog{
			senderID:   make([]MessageAudioDialog, 0, 4),
			receiverID: make([]MessageAudioDialog, 0, 4),
		},
		LastUsedTime: time.Now().UnixNano(),
	}

	s.Mu.Lock()
	s.Dialogs[dialogID] = dialog
	s.Mu.Unlock()

	return
}

func (s *Storage) DeleteDialog(dialogID string) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	delete(s.Dialogs, dialogID)
}

func (s *Storage) ExchangeAudio(idDialog, userID string, message []byte) (map[string][]MessageAudioDialog, error) {
	s.Mu.RLock()
	dialog, ok := s.Dialogs[idDialog]
	s.Mu.RUnlock()
	if !ok {
		return nil, errors.New("dialog not found")
	}

	dialog.Mu.Lock()
	defer dialog.Mu.Unlock()
	if _, exists := dialog.Users[userID]; !exists {
		return nil, errors.New("user not found in dialog")
	}

	now := time.Now().UnixNano()
	dialog.LastUsedTime = now

	for user, messages := range dialog.Users {
		filtered := messages[:0]
		for _, m := range messages {
			if now-m.Time <= timeToSendMessage {
				filtered = append(filtered, m)
			}
		}
		dialog.Users[user] = filtered
	}

	dialog.Users[userID] = append(dialog.Users[userID], MessageAudioDialog{
		Time:  now,
		Audio: message,
	})

	messagesForUser := make(map[string][]MessageAudioDialog)
	for user, messages := range dialog.Users {
		if user == userID {
			continue
		}

		messagesForUser[user] = append([]MessageAudioDialog(nil), messages...)
	}

	for user := range messagesForUser {
		dialog.Users[user] = dialog.Users[user][:0]
	}

	return messagesForUser, nil
}

func (s *Storage) CheckLastUsedTimeInAudioDialog() {
	for {
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

		if len(toDelete) > 0 {
			s.Mu.Lock()
			for _, id := range toDelete {
				d, ok := s.Dialogs[id]
				if !ok {
					continue
				}
				d.Mu.RLock()
				inactive := now-d.LastUsedTime > timeMaxInactive
				d.Mu.RUnlock()
				if inactive {
					delete(s.Dialogs, id)
				}
			}
			s.Mu.Unlock()
		}

		time.Sleep(timeSleepCheckLastUsedTime)
	}
}
