package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gostool/shortid"
)

func main() {
	addr := ":8080"
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	srv, err := shortid.NewHTTPServer(addr, redisAddr, shortid.BusinessOrder)
	if err != nil {
		log.Fatalf("create server failed: %v", err)
	}

	go func() {
		log.Printf("shortid server listening on %s", addr)
		if err := srv.Start(); err != nil {
			log.Printf("server stopped: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown failed: %v", err)
	}
}
