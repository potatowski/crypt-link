//go:build integration

package mongodb

import (
	"context"
	"crypt-link/config"
	"crypt-link/core/domain"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Run with: MONGO_TEST_URL=mongodb://localhost:27017 go test -tags integration ./adapter/output/mongodb/
func setup(t *testing.T) *mongo.Client {
	t.Helper()
	url := os.Getenv("MONGO_TEST_URL")
	if url == "" {
		t.Skip("MONGO_TEST_URL not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		t.Fatalf("ping: %v", err)
	}

	config.DatabaseName = fmt.Sprintf("crypt_link_test_%d", time.Now().UnixNano())
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = client.Database(config.DatabaseName).Drop(ctx)
		_ = client.Disconnect(ctx)
	})
	return client
}

func newMessage(id string) domain.Message {
	now := time.Now().UTC().Truncate(time.Millisecond)
	return domain.Message{ID: id, Encrypted: "cipher", CreatedAt: now, ExpiresAt: now.Add(time.Hour)}
}

func TestSaveAndFindAndDelete(t *testing.T) {
	client := setup(t)
	repo := NewMessageRepository(client)
	ctx := context.Background()

	msg := newMessage("3f2b8c1e-9d4a-4b6f-8e2a-1c5d7f9a0b3e")
	if err := repo.Save(ctx, msg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := repo.FindAndDelete(ctx, msg.ID)
	if err != nil {
		t.Fatalf("FindAndDelete: %v", err)
	}
	if got.Encrypted != msg.Encrypted || !got.ExpiresAt.Equal(msg.ExpiresAt) {
		t.Errorf("got %+v, want %+v", got, msg)
	}

	if _, err := repo.FindAndDelete(ctx, msg.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("second FindAndDelete: got %v, want ErrNotFound", err)
	}
}

func TestSaveDuplicateID(t *testing.T) {
	client := setup(t)
	repo := NewMessageRepository(client)
	ctx := context.Background()

	msg := newMessage("6a1d2f3e-4b5c-4d6e-9f70-8a9b0c1d2e3f")
	if err := repo.Save(ctx, msg); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := repo.Save(ctx, msg); !mongo.IsDuplicateKeyError(err) {
		t.Errorf("duplicate Save: got %v, want duplicate key error", err)
	}
}

func TestConcurrentFindAndDeleteReturnsMessageOnce(t *testing.T) {
	client := setup(t)
	repo := NewMessageRepository(client)
	ctx := context.Background()

	msg := newMessage("0b8e7d6c-5a4f-4e3d-8c2b-1a0f9e8d7c6b")
	if err := repo.Save(ctx, msg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	const n = 32
	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		successes int
		notFound  int
		start     = make(chan struct{})
	)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := repo.FindAndDelete(ctx, msg.ID)
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				successes++
			case errors.Is(err, domain.ErrNotFound):
				notFound++
			default:
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	close(start)
	wg.Wait()

	if successes != 1 || notFound != n-1 {
		t.Fatalf("successes = %d, notFound = %d; want 1 and %d", successes, notFound, n-1)
	}
}

func TestEnsureIndexesIsIdempotent(t *testing.T) {
	client := setup(t)
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		if err := EnsureIndexes(ctx, client); err != nil {
			t.Fatalf("EnsureIndexes call %d: %v", i+1, err)
		}
	}

	cur, err := client.Database(config.DatabaseName).Collection(collectionName).Indexes().List(ctx)
	if err != nil {
		t.Fatalf("list indexes: %v", err)
	}
	var indexes []bson.M
	if err := cur.All(ctx, &indexes); err != nil {
		t.Fatalf("decode indexes: %v", err)
	}
	for _, idx := range indexes {
		if idx["name"] == "expiresAt_ttl" {
			if v, ok := idx["expireAfterSeconds"]; !ok || fmt.Sprint(v) != "0" {
				t.Errorf("expireAfterSeconds = %v, want 0", v)
			}
			return
		}
	}
	t.Fatalf("TTL index not found in %v", indexes)
}
