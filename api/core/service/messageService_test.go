package service

import (
	"context"
	"crypt-link/core/domain"
	"errors"
	"testing"
	"time"
)

// MockMessageRepository is a mock implementation of port.MessageRepository
type MockMessageRepository struct {
	SaveFunc          func(ctx context.Context, msg domain.Message) error
	FindAndDeleteFunc func(ctx context.Context, id string) (*domain.Message, error)
}

func (m *MockMessageRepository) Save(ctx context.Context, msg domain.Message) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, msg)
	}
	return nil
}

func (m *MockMessageRepository) FindAndDelete(ctx context.Context, id string) (*domain.Message, error) {
	if m.FindAndDeleteFunc != nil {
		return m.FindAndDeleteFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

var fixedNow = time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

func fixedClock() time.Time { return fixedNow }

func TestCreate(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockMessageRepository{}
	svc := NewMessageService(mockRepo, fixedClock)

	t.Run("fails when expiration is in the past", func(t *testing.T) {
		err := svc.Create(ctx, "id-123", "secret", fixedNow.Add(-1*time.Hour))
		if !errors.Is(err, domain.ErrInvalidInput) {
			t.Errorf("Expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("fails when id or encrypted are empty", func(t *testing.T) {
		err := svc.Create(ctx, "", "secret", fixedNow.Add(1*time.Hour))
		if !errors.Is(err, domain.ErrInvalidInput) {
			t.Errorf("Expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("successfully creates a message", func(t *testing.T) {
		expiresAt := fixedNow.Add(1 * time.Hour)
		mockRepo.SaveFunc = func(_ context.Context, msg domain.Message) error {
			if msg.ID != "123" {
				t.Errorf("Expected ID 123, got %s", msg.ID)
			}
			if !msg.CreatedAt.Equal(fixedNow) {
				t.Errorf("Expected CreatedAt %v, got %v", fixedNow, msg.CreatedAt)
			}
			if !msg.ExpiresAt.Equal(expiresAt) {
				t.Errorf("Expected ExpiresAt %v, got %v", expiresAt, msg.ExpiresAt)
			}
			return nil
		}
		err := svc.Create(ctx, "123", "secret", expiresAt)
		if err != nil {
			t.Errorf("Expected success, got %v", err)
		}
	})

	t.Run("propagates repository errors", func(t *testing.T) {
		repoErr := errors.New("db down")
		mockRepo.SaveFunc = func(context.Context, domain.Message) error { return repoErr }
		err := svc.Create(ctx, "123", "secret", fixedNow.Add(1*time.Hour))
		if !errors.Is(err, repoErr) {
			t.Errorf("Expected repository error, got %v", err)
		}
	})
}

func TestGetAndInvalidate(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockMessageRepository{}
	svc := NewMessageService(mockRepo, fixedClock)

	t.Run("fails when message is not found", func(t *testing.T) {
		mockRepo.FindAndDeleteFunc = func(context.Context, string) (*domain.Message, error) {
			return nil, domain.ErrNotFound
		}
		_, err := svc.GetAndInvalidate(ctx, "123")
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})

	t.Run("fails when message is expired", func(t *testing.T) {
		mockRepo.FindAndDeleteFunc = func(context.Context, string) (*domain.Message, error) {
			return &domain.Message{
				ID:        "123",
				ExpiresAt: fixedNow.Add(-1 * time.Second),
			}, nil
		}
		_, err := svc.GetAndInvalidate(ctx, "123")
		if !errors.Is(err, domain.ErrExpired) {
			t.Errorf("Expected ErrExpired, got %v", err)
		}
	})

	t.Run("successfully retrieves message", func(t *testing.T) {
		mockRepo.FindAndDeleteFunc = func(context.Context, string) (*domain.Message, error) {
			return &domain.Message{
				ID:        "123",
				Encrypted: "encrypted",
				ExpiresAt: fixedNow.Add(1 * time.Hour),
			}, nil
		}
		msg, err := svc.GetAndInvalidate(ctx, "123")
		if err != nil {
			t.Fatalf("Expected success, got %v", err)
		}
		if msg.Encrypted != "encrypted" {
			t.Errorf("Expected encrypted, got %s", msg.Encrypted)
		}
	})
}

func TestMessageIsExpired(t *testing.T) {
	m := domain.Message{ExpiresAt: fixedNow}
	if m.IsExpired(fixedNow) {
		t.Error("message should not be expired at its expiration instant")
	}
	if !m.IsExpired(fixedNow.Add(time.Nanosecond)) {
		t.Error("message should be expired after its expiration instant")
	}
}
