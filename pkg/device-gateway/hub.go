package gateway

import (
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type hub struct {
	connections map[*hubConnection]bool

	// sub requests from the clients.
	sub chan *hubConnection

	// Unsub requests from clients.
	unsub chan *hubConnection

	// Message to broadcast
	in chan string
}

func newHub() *hub {
	return &hub{
		connections: make(map[*hubConnection]bool),
		sub:         make(chan *hubConnection),
		unsub:       make(chan *hubConnection),
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
				c.send <- message
			}
		}
	}
}

type hubConnection struct {
	*websocket.Conn

	send chan interface{}
}
