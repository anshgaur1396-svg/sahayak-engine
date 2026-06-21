# Sahayak: Self-Healing Transaction Assurance Engine

[![Go Reference](https://pkg.go.dev/badge/net/http.svg)](https://pkg.go.dev/net/http)
[![React](https://img.shields.io/badge/React-18-blue.svg)](https://reactjs.org/)

**Sahayak** is a deterministic, highly-concurrent microservice designed to eliminate the "Debited-But-Uncredited" Panic Cycle in digital banking. By combining cryptographic state locking with an asynchronous self-healing worker pool, Sahayak protects user funds and drastically reduces support center loads.

## The Three Pillars of Sahayak

1. **Idempotent Transaction Guard:** A custom Go middleware mathematically prevents duplicate executions (double-debits) using `X-Idempotency-Key` headers, even under heavy UI spam or network latency.
2. **Proactive Panic Lock:** A real-time WebSocket broadcaster detects gateway timeouts (HTTP 504) and instantly freezes the client UI, communicating certainty to the user ("Funds are safe. Retrying...").
3. **Asynchronous Self-Healing Queue:** Failed transactions do not crash the session. They are routed to a concurrent Go worker pool executing a systematic back-off retry loop, delivering a final success state without user intervention.

## System Architecture

```text
[Client UI] (React/Vite)
    │    ▲
    │    │ (WebSocket State Broadcasts)
    ▼    │
[ Go Backend Engine ] ──(CORS & Security Middleware)──┐
    │                                                 │
    ├─► [Idempotency Middleware] ──(Checks Ledger)──► [Redis/In-Memory State]
    │
    └─► [Transaction Handler] ──(Simulated 504)
            │
            ▼
      [ Async Worker Pool (Goroutines) ]
            │ (Systematic Back-off Retry)
            ▼
      [ Final Resolution / Rollback ]
