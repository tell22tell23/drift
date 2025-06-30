package api

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sammanbajracharya/drift/internal/utils"
)

type SignalingMessage struct {
	logger *log.Logger
}

type Client struct {
	conn *websocket.Conn
	send chan []byte
	dft  string
}

var (
	dfts     = make(map[string]map[*Client]bool)
	dftsMu   sync.Mutex
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func NewSignalingMessage(logger *log.Logger) *SignalingMessage {
	return &SignalingMessage{
		logger: logger,
	}
}

func (sm *SignalingMessage) HandleSignaling(w http.ResponseWriter, r *http.Request) {
	dftID := r.URL.Query().Get("dft_id")
	if dftID == "" {
		http.Error(w, "dft_id is required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		sm.logger.Printf("WebSocket upgrade failed: %v\n", err)
		return
	}

	client := &Client{
		conn: conn,
		send: make(chan []byte, 256),
		dft:  dftID,
	}

	dftsMu.Lock()
	if dfts[dftID] == nil {
		dfts[dftID] = make(map[*Client]bool)
	}
	dfts[dftID][client] = true
	dftsMu.Unlock()

	go func() {
		for msg := range client.send {
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				sm.logger.Printf("Error sending message: %v\n", err)
				break
			}
		}
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			sm.logger.Printf("Error reading message: %v\n", err)
			break
		}

		dftsMu.Lock()
		for c := range dfts[dftID] {
			if c != client {
				select {
				case c.send <- msg:
				default:
					sm.logger.Printf("Client send channel full, dropping message\n")
				}
			}
		}
		dftsMu.Unlock()
	}

	dftsMu.Lock()
	delete(dfts[dftID], client)
	if len(dfts[dftID]) == 0 {
		delete(dfts, dftID)
	}
	dftsMu.Unlock()

	close(client.send)
	conn.Close()

	sm.logger.Printf("Client disconnected from dft %s", dftID)
}

func (sm *SignalingMessage) HandleStats(w http.ResponseWriter, r *http.Request) {
	dftsMu.Lock()
	defer dftsMu.Unlock()

	dftID := r.URL.Query().Get("dft_id")
	if dftID == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "dft_id is required"})
		return
	}

	count := len(dfts[dftID])
	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"total": count})
}
