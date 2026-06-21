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

// corsMiddleware intercepts all requests, handling preflight checks and granting port access
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow requests from any port
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "X-Idempotency-Key, Content-Type")

		// Instantly approve the browser's hidden preflight request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	InitRedis()
	
	go StartWorkerPool(5) 

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", WebSocketHandler)
	mux.Handle("/transaction", IdempotencyMiddleware(http.HandlerFunc(TransactionHandler)))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: corsMiddleware(mux), // Wrap the entire engine router in our new CORS guard
	}

	go func() {
		log.Println("Sahayak Engine active on port 8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

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