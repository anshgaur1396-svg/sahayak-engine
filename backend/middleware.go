package main

import (
	"encoding/json"
	"net/http"
	"sync"
)

// In-memory ledger simulating Redis for MVP
var redisMock = make(map[string]bool)
var mu sync.Mutex

func InitRedis() {} // Placeholder

func IdempotencyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idemKey := r.Header.Get("X-Idempotency-Key")
		if idemKey == "" {
			http.Error(w, "Idempotency key required", http.StatusBadRequest)
			return
		}

		mu.Lock()
		if redisMock[idemKey] {
			mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict) // 409 Conflict
			json.NewEncoder(w).Encode(map[string]string{"status": "DUPLICATE_PREVENTED", "message": "Transaction already locked."})
			return
		}
		redisMock[idemKey] = true
		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}