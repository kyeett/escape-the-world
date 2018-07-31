package gateway

import (
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type hub struct {
	connections map[chan string]bool

	// sub requests from the clients.
	sub chan chan string

	// Unsub requests from clients.
	unsub chan chan string

	// Message to broadcast
	in chan string
}

func newHub() *hub {
	return &hub{
		connections: make(map[chan string]bool),
		sub:         make(chan chan string),
		unsub:       make(chan chan string),
		in:          make(chan string),
	}
}

func (h *hub) run() {
	for {
		select {
		case conn := <-h.sub:
			logrus.Info("Sub")
			h.connections[conn] = true

		case conn := <-h.unsub:
			logrus.Info("Unsub")
			delete(h.connections, conn)

		case message := <-h.in:
			logrus.Info("Send msg")
			for c := range h.connections {
				c <- message
			}
		}
	}
}

type hubConnection struct {
	*websocket.Conn

	send chan interface{}
}
