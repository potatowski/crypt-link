//go:build js && wasm

// Command worker runs the API on Cloudflare Workers with a D1 database.
package main

import (
	"context"
	"crypt-link/adapter/input/controller"
	"crypt-link/adapter/output/sqlstore"
	"crypt-link/core/service"
	"crypt-link/infra/middleware"
	"crypt-link/infra/router"
	"database/sql"
	"log/slog"
	"time"

	"github.com/syumai/workers-go"
	"github.com/syumai/workers-go/cloudflare/cron"
	_ "github.com/syumai/workers-go/cloudflare/d1"
)

// d1Binding is the D1 binding name declared in wrangler.toml.
const d1Binding = "DB"

func main() {
	db, err := sql.Open("d1", d1Binding)
	if err != nil {
		slog.Error("failed to open D1 database", "error", err)
		return
	}

	repo := sqlstore.NewMessageRepository(db)
	ctrl := controller.NewMessageController(service.NewMessageService(repo, nil))

	workers.ServeNonBlock(middleware.EdgeHeaders(router.SetupWithoutRateLimit(ctrl)))
	cron.ScheduleTaskNonBlock(func(ctx context.Context) error {
		n, err := repo.DeleteExpired(ctx, time.Now().UTC())
		if err != nil {
			// workers-go panics when a scheduled task returns an error; log it
			// instead; the next hourly run retries.
			slog.Error("failed to delete expired messages", "error", err)
			return nil
		}
		slog.Info("expired messages deleted", "count", n)
		return nil
	})
	workers.Ready()

	select {
	case <-workers.Done():
	case <-cron.Done():
	}
}
