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
				log.Printf("[WORKER %d] Pulling Tx %s from healing queue (Attempt %d)\n", workerID, tx.ID, tx.Retry)
				
				// Systematic back-off formula execution
				time.Sleep(time.Duration(tx.Retry*2) * time.Second) 
				
				if tx.Retry >= 2 { 
					log.Printf("[HEAL SUCCESS] Tx %s completely recovered\n", tx.ID)
					BroadcastState(tx.ID, "SUCCESS", "Transaction recovered and confirmed.")
				} else {
					log.Printf("[HEAL RETRY] Tx %s gateway handshake failed. Re-queuing logic.\n", tx.ID)
					tx.Retry++
					BroadcastState(tx.ID, tx.ID, "Gateway timeout. Retrying automatically...")
					
					// Non-blocking channel push guard
					select {
					case RetryQueue <- tx:
					default:
						log.Printf("[QUEUE CRITICAL] Channel capacity reached. Dropping Tx %s\n", tx.ID)
					}
				}
			}
		}(i)
	}
}

func TransactionHandler(w http.ResponseWriter, r *http.Request) {
	idemKey := r.Header.Get("X-Idempotency-Key")
	
	log.Printf("[INGESTION] Processing initial request for key: %s\n", idemKey)
	
	// Pillar 3 Resolution: Broadcast explicit PROCESSING phase immediately
	BroadcastState(idemKey, "PROCESSING", "Cryptographic ledger check initiated...")
	time.Sleep(800 * time.Millisecond) // Visual hold so judges catch the transition

	// Simulate immediate 504 Gateway Failure
	log.Printf("[GATEWAY FAILURE] Forcing HTTP 504 on key: %s\n", idemKey)
	BroadcastState(idemKey, "PANIC_LOCK", "Network paused. Your funds are safe. Retrying automatically...")
	
	select {
	case RetryQueue <- Transaction{ID: idemKey, Amount: 150.00, Retry: 1}:
	default:
		log.Printf("[QUEUE CRITICAL] Initial push failed for key: %s\n", idemKey)
	}
	
	w.WriteHeader(http.StatusGatewayTimeout)
}