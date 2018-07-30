package gateway

import (
	"encoding/json"
	"errors"
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
		return errors.New("device already registered")
	}

	devices[info.ID] = info
	return nil
}

func disconnectDevice(id string) {
	dMutex.Lock()
	d := devices[id]
	d.Connected = false
	devices[id] = d
	dMutex.Unlock()
}

func listDevices(w http.ResponseWriter, r *http.Request) {
	dMutex.RLock()
	defer dMutex.RUnlock()

	b, err := json.MarshalIndent(devices, "", "   ")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	logrus.Errorf("%v", string(b))
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
}

func Start(srv *http.Server) {

	http.HandleFunc("/connect", connect)
	http.HandleFunc("/", listDevices)
	logrus.Fatal(srv.ListenAndServe())
}
