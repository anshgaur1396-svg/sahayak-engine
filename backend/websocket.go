package main

import (
	"log"
	"net/http"
	"sync"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var clients = make(map[*websocket.Conn]bool)
var wsMutex sync.Mutex

type WSMessage struct {
	TxID    string `json:"tx_id"`
	State   string `json:"state"` // IDLE, PROCESSING, PANIC_LOCK, RETRYING, SUCCESS
	Message string `json:"message"`
}

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS ERROR] Upgrade failed: %v\n", err)
		return
	}
	
	wsMutex.Lock()
	clients[ws] = true
	wsMutex.Unlock()
	log.Printf("[WS CONNECT] New client established memory pointer: %p\n", ws)

	// Concurrent Read Loop: Detects client disconnects/refreshes instantly
	go func(conn *websocket.Conn) {
		defer func() {
			wsMutex.Lock()
			delete(clients, conn)
			wsMutex.Unlock()
			conn.Close()
			log.Printf("[WS DISCONNECT] Safely evicted dead client: %p\n", conn)
		}()

		for {
			// Keeps connection alive and reads incoming frames; fails immediately if browser closes/refreshes
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}(ws)
}

func BroadcastState(txID, state, message string) {
	wsMutex.Lock()
	defer wsMutex.Unlock()

	log.Printf("[STATE BROADCAST] Code: %-12s | Tx: %s\n", state, txID)
	msg := WSMessage{TxID: txID, State: state, Message: message}

	for client := range clients {
		err := client.WriteJSON(msg)
		if err != nil {
			log.Printf("[WS WRITE ERROR] Failed on %p, evicting reference: %v\n", client, err)
			client.Close()
			delete(clients, client) // Immediate runtime cleanup
		}
	}
}