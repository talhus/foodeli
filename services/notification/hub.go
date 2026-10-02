package notification

import (
	"sync"

	"github.com/google/uuid"
)

type Hub struct {
	mu      sync.RWMutex
	clients map[uuid.UUID]map[chan []byte]bool
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[uuid.UUID]map[chan []byte]bool),
	}
}

func (h *Hub) Register(userID uuid.UUID) chan []byte {
	h.mu.Lock()
	defer h.mu.Unlock()
	ch := make(chan []byte, 16)
	if _, exist := h.clients[userID]; !exist {
		h.clients[userID] = make(map[chan []byte]bool)
	}
	h.clients[userID][ch] = true

	return ch
}

func (h *Hub) Unregister(userID uuid.UUID, ch chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, exist := h.clients[userID]; exist {
		if _, ok := h.clients[userID][ch]; ok {
			close(ch)
			delete(h.clients[userID], ch)
		}
		if len(h.clients[userID]) == 0 {
			delete(h.clients, userID)
		}
	}
}

func (h *Hub) SendToUser(userID uuid.UUID, data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for ch := range h.clients[userID] {

		select {
		case ch <- data:
		default:
		}

	}

}
