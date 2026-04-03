package service

import (
	"crypt-link/core/domain"
	"errors"
	"testing"
	"time"
)

// MockMessageRepository is a mock implementation of port.MessageRepository
type MockMessageRepository struct {
	SaveFunc          func(msg domain.Message) error
	FindAndDeleteFunc func(id string) (*domain.Message, error)
}

func (m *MockMessageRepository) Save(msg domain.Message) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(msg)
	}
	return nil
}

func (m *MockMessageRepository) FindAndDelete(id string) (*domain.Message, error) {
	if m.FindAndDeleteFunc != nil {
		return m.FindAndDeleteFunc(id)
	}
	return nil, errors.New("not implemented")
}

func TestCreate(t *testing.T) {
	mockRepo := &MockMessageRepository{}
	svc := NewMessageService(mockRepo)

	t.Run("fails when expiration is in the past", func(t *testing.T) {
		err := svc.Create("id-123", "secret", time.Now().Add(-1*time.Hour))
		if err == nil || err.Error() != "expiration time must be in the future" {
			t.Errorf("Expected error expiration in the past, got %v", err)
		}
	})

	t.Run("fails when id or encrypted are empty", func(t *testing.T) {
		err := svc.Create("", "secret", time.Now().Add(1*time.Hour))
		if err == nil || err.Error() != "id and encrypted content must not be empty" {
			t.Errorf("Expected error empty fields, got %v", err)
		}
	})

	t.Run("successfully creates a message", func(t *testing.T) {
		mockRepo.SaveFunc = func(msg domain.Message) error {
			if msg.ID != "123" {
				t.Errorf("Expected ID 123, got %s", msg.ID)
			}
			return nil
		}
		err := svc.Create("123", "secret", time.Now().Add(1*time.Hour))
		if err != nil {
			t.Errorf("Expected success, got %v", err)
		}
	})
}

func TestGetAndInvalidate(t *testing.T) {
	mockRepo := &MockMessageRepository{}
	svc := NewMessageService(mockRepo)

	t.Run("fails when message is not found", func(t *testing.T) {
		mockRepo.FindAndDeleteFunc = func(id string) (*domain.Message, error) {
			return nil, errors.New("not found")
		}
		_, err := svc.GetAndInvalidate("123")
		if err == nil || err.Error() != "not found" {
			t.Errorf("Expected not found error, got %v", err)
		}
	})

	t.Run("fails when message is expired", func(t *testing.T) {
		mockRepo.FindAndDeleteFunc = func(id string) (*domain.Message, error) {
			return &domain.Message{
				ID:        "123",
				ExpiresAt: time.Now().Add(-1 * time.Hour),
			}, nil
		}
		_, err := svc.GetAndInvalidate("123")
		if err == nil || err.Error() != "expired" {
			t.Errorf("Expected expired error, got %v", err)
		}
	})

	t.Run("successfully retrieves message", func(t *testing.T) {
		mockRepo.FindAndDeleteFunc = func(id string) (*domain.Message, error) {
			return &domain.Message{
				ID:        "123",
				Encrypted: "encrypted",
				ExpiresAt: time.Now().Add(1 * time.Hour),
			}, nil
		}
		msg, err := svc.GetAndInvalidate("123")
		if err != nil {
			t.Errorf("Expected success, got %v", err)
		}
		if msg.Encrypted != "encrypted" {
			t.Errorf("Expected encrypted, got %s", msg.Encrypted)
		}
	})
}
