package sqlstore

import (
	"context"
	"crypt-link/core/domain"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func newRepo(t *testing.T) (*MessageRepository, *sql.DB) {
	t.Helper()
	dsn := "file:" + filepath.Join(t.TempDir(), "test.db") + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.Exec(Schema); err != nil {
		t.Fatal(err)
	}
	return NewMessageRepository(db), db
}

func message(id string, expiresIn time.Duration) domain.Message {
	now := time.Now().UTC().Truncate(time.Millisecond)
	return domain.Message{ID: id, Encrypted: "cipher-" + id, CreatedAt: now, ExpiresAt: now.Add(expiresIn)}
}

func TestSaveAndFindAndDelete(t *testing.T) {
	repo, _ := newRepo(t)
	ctx := context.Background()
	want := message("3f2b8c1e-9d4a-4b6f-8e2a-1c5d7f9a0b3e", time.Hour)

	if err := repo.Save(ctx, want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := repo.FindAndDelete(ctx, want.ID)
	if err != nil {
		t.Fatalf("FindAndDelete: %v", err)
	}
	if got.ID != want.ID || got.Encrypted != want.Encrypted || !got.CreatedAt.Equal(want.CreatedAt) || !got.ExpiresAt.Equal(want.ExpiresAt) {
		t.Fatalf("got %+v, want %+v", *got, want)
	}

	if _, err := repo.FindAndDelete(ctx, want.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("second FindAndDelete error = %v, want domain.ErrNotFound", err)
	}
}

func TestFindAndDeleteUnknown(t *testing.T) {
	repo, _ := newRepo(t)
	if _, err := repo.FindAndDelete(context.Background(), "missing"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("error = %v, want domain.ErrNotFound", err)
	}
}

func TestSaveDuplicateID(t *testing.T) {
	repo, _ := newRepo(t)
	ctx := context.Background()
	msg := message("dup", time.Hour)
	if err := repo.Save(ctx, msg); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(ctx, msg); err == nil {
		t.Fatal("saving a duplicate id succeeded, want an error")
	}
}

func TestFindAndDeleteReturnsMessageOnce(t *testing.T) {
	repo, _ := newRepo(t)
	ctx := context.Background()
	msg := message("concurrent", time.Hour)
	if err := repo.Save(ctx, msg); err != nil {
		t.Fatal(err)
	}

	const readers = 20
	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		found    int
		notFound int
	)
	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := repo.FindAndDelete(ctx, msg.ID)
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				found++
			case errors.Is(err, domain.ErrNotFound):
				notFound++
			default:
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()
	if found != 1 || notFound != readers-1 {
		t.Fatalf("found = %d, notFound = %d; want 1 and %d", found, notFound, readers-1)
	}
}

func TestDeleteExpired(t *testing.T) {
	repo, db := newRepo(t)
	ctx := context.Background()
	for _, m := range []domain.Message{message("expired", -time.Minute), message("live", time.Hour)} {
		if err := repo.Save(ctx, m); err != nil {
			t.Fatal(err)
		}
	}

	n, err := repo.DeleteExpired(ctx, time.Now().UTC())
	if err != nil {
		t.Fatalf("DeleteExpired: %v", err)
	}
	if n != 1 {
		t.Fatalf("deleted = %d, want 1", n)
	}
	var ids []string
	rows, err := db.Query(`SELECT id FROM messages`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if strings.Join(ids, ",") != "live" {
		t.Fatalf("remaining ids = %v, want [live]", ids)
	}
}

// The D1 migration must create the same schema the repository expects.
func TestMigrationMatchesSchema(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "worker", "migrations", "0001_create_messages.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(b)) != strings.TrimSpace(Schema) {
		t.Fatalf("worker/migrations/0001_create_messages.sql differs from sqlstore.Schema")
	}
}
