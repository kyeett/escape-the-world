package gateway

import (
	"encoding/json"
	"errors"
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

// DeviceInfo holds basic information about a device, and whether it is connected
type DeviceInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	Description string `json:"description"`
	Connected   bool   `json:"connected"`
}

type deviceList map[string]DeviceInfo

var (
	dMutex  = &sync.RWMutex{}
	devices = deviceList{}
)

func registerDevice(info DeviceInfo) error {
	dMutex.Lock()
	defer dMutex.Unlock()
	if info.ID == "" {
		return errors.New("no device ID specified")
	}

	if _, alreadyExists := devices[info.ID]; alreadyExists {
		logrus.Error("device already registered")
	}

	info.Connected = true
	devices[info.ID] = info

	h.in <- ""
	return nil
}

var disconnectDoneTestHook = func() {}

func disconnectDevice(id string) {

	dMutex.Lock()
	d := devices[id]
	d.Connected = false
	devices[id] = d
	dMutex.Unlock()
	logrus.Infof("device '%s' disconnected", id)

	h.in <- ""
	disconnectDoneTestHook()
}

func listDevices(w http.ResponseWriter, r *http.Request) {
	dMutex.RLock()
	defer dMutex.RUnlock()
	di := []DeviceInfo{}
	for key := range devices {
		di = append(di, devices[key])
	}

	b, err := json.MarshalIndent(di, "", "   ")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write(b)

}

func connect(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Error("ws upgrade:", err)
		return
	}
	defer c.Close()

	var message DeviceInfo
	err = c.ReadJSON(&message)
	if err != nil {
		logrus.Error("ws read:", err)
		return
	}

	if err := registerDevice(message); err != nil {
		logrus.Error("register device:", err)
		return
	}
	logrus.Infof("device connected: %v", devices)
	defer disconnectDevice(message.ID)

	response := map[string]string{
		"response": "OK",
	}

	if err := c.WriteJSON(response); err != nil {
		logrus.Error("ws write:", err)
		return
	}
	for {
		logrus.Info("waiting for updates")
		var update interface{}
		err := c.ReadJSON(update)
		if err != nil {
			logrus.Error("ws read:", err)
			return
		}
	}
}

func handleBrowserConnections(conn *websocket.Conn) {
	send := make(chan interface{})
	wsConn := hubConnection{
		conn,
		send,
	}

	h.sub <- &wsConn
	defer func() {
		h.unsub <- &wsConn
		close(wsConn.send)
	}()

	for {
		<-wsConn.send
		m := struct {
			Message string
		}{
			"updated",
		}
		if err := conn.WriteJSON(m); err != nil {
			fmt.Println(err)
			break
		}
	}
}

var h *hub

func Start(srv *http.Server) {
	h = newHub()
	go h.run()

	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/list", listDevices)
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/connect", connect)
	logrus.Fatal(srv.ListenAndServe())
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Origin") != "http://"+r.Host {
		http.Error(w, "Origin not allowed", 403)
		return
	}
	conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}
	defer conn.Close()

	handleBrowserConnections(conn)

}
