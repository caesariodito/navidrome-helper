package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"navidrome-helper/internal/config"
	"navidrome-helper/internal/jobs"
	"navidrome-helper/internal/library"
	"navidrome-helper/internal/server"
	"navidrome-helper/internal/store"
)

func main() {
	cfg := config.Load()
	dbPath := filepath.Join(cfg.DataDir, "navidrome-helper.db")
	store, err := store.New(dbPath)
	if err != nil {
		log.Fatalf("init store: %v", err)
	}

	runner := jobs.NewRunner(store, cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runner.Start(ctx)

	indexer := library.NewIndexer(cfg, store)
	if _, err := indexer.Refresh(ctx); err != nil {
		log.Printf("library refresh at start failed: %v", err)
	}

	srv := server.New(cfg, store, runner, indexer)
	go func() {
		log.Printf("backend listening on :%s", cfg.Port)
		if err := http.ListenAndServe(":"+cfg.Port, srv.Routes()); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Wait for interrupt
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	log.Println("shutting down...")
	cancel()
}
