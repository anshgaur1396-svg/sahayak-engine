package main

import (
	"log"
	"net/http"
	"time"
)

type Transaction struct {
	ID     string
	Amount float64
	Retry  int
}

var RetryQueue = make(chan Transaction, 100)

func StartWorkerPool(workers int) {
	for i := 1; i <= workers; i++ {
		go func(workerID int) {
			for tx := range RetryQueue {
				log.Printf("Worker %d processing Tx %s (Retry %d)\n", workerID, tx.ID, tx.Retry)
				
				// Simulate systematic back-off
				time.Sleep(time.Duration(tx.Retry*2) * time.Second) 
				
				// 80% chance of success on retry
				if tx.Retry >= 2 { 
					log.Printf("Tx %s Recovered successfully.", tx.ID)
					BroadcastState(tx.ID, "SUCCESS", "Transaction recovered and confirmed.")
				} else {
					log.Printf("Tx %s failed. Re-queuing.", tx.ID)
					BroadcastState(tx.ID, "RETRYING", "Gateway timeout. Retrying automatically...")
					tx.Retry++
					RetryQueue <- tx
				}
			}
		}(i)
	}
}

// TransactionHandler simulates the primary failure
func TransactionHandler(w http.ResponseWriter, r *http.Request) {
	idemKey := r.Header.Get("X-Idempotency-Key")
	
	// Simulate immediate 504 Gateway Timeout
	BroadcastState(idemKey, "PANIC_LOCK", "Network paused. Your funds are safe. Retrying automatically...")
	
	// Route to background healing queue
	RetryQueue <- Transaction{ID: idemKey, Amount: 150.00, Retry: 1}
	
	w.WriteHeader(http.StatusGatewayTimeout)
}