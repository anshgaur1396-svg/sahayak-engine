package main

import (
	"log"
	"net/http"
	"sync"
	"github.com/gorilla/websocket" // Ensure to go get this
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
		log.Fatal(err)
	}
	wsMutex.Lock()
	clients[ws] = true
	wsMutex.Unlock()
}

func BroadcastState(txID, state, message string) {
	wsMutex.Lock()
	defer wsMutex.Unlock()
	msg := WSMessage{TxID: txID, State: state, Message: message}
	for client := range clients {
		client.WriteJSON(msg)
	}
}