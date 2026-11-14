package service

import (
	"context"
	"crypt-link/core/domain"
	"errors"
	"strings"
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

const testID = "3f2b8c1e-9d4a-4b6f-8e2a-1c5d7f9a0b3e"

var fixedNow = time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

func fixedClock() time.Time { return fixedNow }

func TestCreate(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockMessageRepository{}
	svc := NewMessageService(mockRepo, fixedClock)

	t.Run("fails when expiresIn is not an allowed value", func(t *testing.T) {
		for _, secs := range []int64{-1, 1, 3599, 2592000} {
			err := svc.Create(ctx, testID, "secret", secs)
			if !errors.Is(err, domain.ErrInvalidInput) {
				t.Errorf("expiresIn %d: expected ErrInvalidInput, got %v", secs, err)
			}
		}
	})

	t.Run("fails when id or encrypted are empty", func(t *testing.T) {
		err := svc.Create(ctx, "", "secret", 0)
		if !errors.Is(err, domain.ErrInvalidInput) {
			t.Errorf("Expected ErrInvalidInput, got %v", err)
		}
		err = svc.Create(ctx, testID, "", 0)
		if !errors.Is(err, domain.ErrInvalidInput) {
			t.Errorf("Expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("fails when id is not a UUID v4", func(t *testing.T) {
		for _, id := range []string{"not-a-uuid", "3f2b8c1e-9d4a-1b6f-8e2a-1c5d7f9a0b3e", "3f2b8c1e9d4a4b6f8e2a1c5d7f9a0b3e"} {
			err := svc.Create(ctx, id, "secret", 0)
			if !errors.Is(err, domain.ErrInvalidInput) {
				t.Errorf("id %q: expected ErrInvalidInput, got %v", id, err)
			}
		}
	})

	t.Run("fails when encrypted content is too long", func(t *testing.T) {
		long := strings.Repeat("a", domain.MaxEncryptedLength+1)
		err := svc.Create(ctx, testID, long, 0)
		if !errors.Is(err, domain.ErrInvalidInput) {
			t.Errorf("Expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("successfully creates a message", func(t *testing.T) {
		expiresAt := fixedNow.Add(1 * time.Hour)
		mockRepo.SaveFunc = func(_ context.Context, msg domain.Message) error {
			if msg.ID != testID {
				t.Errorf("Expected ID %s, got %s", testID, msg.ID)
			}
			if !msg.CreatedAt.Equal(fixedNow) {
				t.Errorf("Expected CreatedAt %v, got %v", fixedNow, msg.CreatedAt)
			}
			if !msg.ExpiresAt.Equal(expiresAt) {
				t.Errorf("Expected ExpiresAt %v, got %v", expiresAt, msg.ExpiresAt)
			}
			return nil
		}
		err := svc.Create(ctx, testID, "secret", 3600)
		if err != nil {
			t.Errorf("Expected success, got %v", err)
		}
	})

	t.Run("uses default expiry when expiresIn is omitted", func(t *testing.T) {
		for secs, want := range map[int64]time.Duration{0: domain.DefaultExpiry, 86400: 24 * time.Hour, 604800: 7 * 24 * time.Hour} {
			mockRepo.SaveFunc = func(_ context.Context, msg domain.Message) error {
				if got := msg.ExpiresAt.Sub(fixedNow); got != want {
					t.Errorf("expiresIn %d: lifetime = %v, want %v", secs, got, want)
				}
				return nil
			}
			if err := svc.Create(ctx, testID, "secret", secs); err != nil {
				t.Errorf("expiresIn %d: expected success, got %v", secs, err)
			}
		}
	})

	t.Run("propagates repository errors", func(t *testing.T) {
		repoErr := errors.New("db down")
		mockRepo.SaveFunc = func(context.Context, domain.Message) error { return repoErr }
		err := svc.Create(ctx, testID, "secret", 0)
		if !errors.Is(err, repoErr) {
			t.Errorf("Expected repository error, got %v", err)
		}
	})
}

func TestGetAndInvalidate(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockMessageRepository{}
	svc := NewMessageService(mockRepo, fixedClock)

	t.Run("fails when id is invalid without hitting the repository", func(t *testing.T) {
		mockRepo.FindAndDeleteFunc = func(context.Context, string) (*domain.Message, error) {
			t.Error("repository should not be called")
			return nil, nil
		}
		_, err := svc.GetAndInvalidate(ctx, "not-a-uuid")
		if !errors.Is(err, domain.ErrInvalidInput) {
			t.Errorf("Expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("fails when message is not found", func(t *testing.T) {
		mockRepo.FindAndDeleteFunc = func(context.Context, string) (*domain.Message, error) {
			return nil, domain.ErrNotFound
		}
		_, err := svc.GetAndInvalidate(ctx, testID)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})

	t.Run("fails when message is expired", func(t *testing.T) {
		mockRepo.FindAndDeleteFunc = func(context.Context, string) (*domain.Message, error) {
			return &domain.Message{
				ID:        testID,
				ExpiresAt: fixedNow.Add(-1 * time.Second),
			}, nil
		}
		_, err := svc.GetAndInvalidate(ctx, testID)
		if !errors.Is(err, domain.ErrExpired) {
			t.Errorf("Expected ErrExpired, got %v", err)
		}
	})

	t.Run("successfully retrieves message", func(t *testing.T) {
		mockRepo.FindAndDeleteFunc = func(context.Context, string) (*domain.Message, error) {
			return &domain.Message{
				ID:        testID,
				Encrypted: "encrypted",
				ExpiresAt: fixedNow.Add(1 * time.Hour),
			}, nil
		}
		msg, err := svc.GetAndInvalidate(ctx, testID)
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
