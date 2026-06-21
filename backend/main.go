package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Initialize Redis (Mocked here for MVP, replace with go-redis)
	InitRedis()
	
	// Start Worker Pool
	go StartWorkerPool(5) 

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", WebSocketHandler)
	mux.Handle("/transaction", IdempotencyMiddleware(http.HandlerFunc(TransactionHandler)))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Server execution in a goroutine
	go func() {
		log.Println("Sahayak Engine active on port 8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful Shutdown OS Signal Trap
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Sahayak Engine gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Engine forced to shutdown: %v", err)
	}
	log.Println("Engine stopped. All states secured.")
}