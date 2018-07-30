package gateway

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
var status = 0
var statusText = map[int]string{
	0: "OFF",
	1: "ON",
}
var mutex sync.RWMutex

func connect(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Error("ws upgrade:", err)
		return
	}
	defer c.Close()
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			logrus.Error("ws read:", err)
			break
		}
		logrus.Infof("ws recv: %s", string(message))

		response := map[string]string{
			"response": "OK",
		}

		if err := c.WriteJSON(response); err != nil {
			logrus.Error("ws write:", err)
			break
		}
	}
}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Error("ws upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			logrus.Error("ws read:", err)
			break
		}
		logrus.Infof("ws recv: %s", string(message))
		err = c.WriteMessage(mt, append(message, byte('\n')))
		if err != nil {
			logrus.Error("ws write:", err)
			break
		}
	}
}

func Start(srv *http.Server) {

	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		mutex.RLock()
		msg := fmt.Sprintf("000%d", status)
		mutex.RUnlock()
		logrus.Info("read:", msg)
		fmt.Fprint(w, msg)
	})

	http.HandleFunc("/toggle", func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		status = 1 - status
		msg := fmt.Sprintf("Led turned %s\t(000%d)", statusText[status], status)
		mutex.Unlock()
		logrus.Info("toggle:", msg)
		fmt.Fprint(w, msg)
	})

	http.HandleFunc("/echo", echo)
	http.HandleFunc("/connect", connect)

	if err := srv.ListenAndServe(); err != nil {
		logrus.Error(err)
	}
}
