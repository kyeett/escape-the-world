package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/connect"}

	deviceInfo := map[string]interface{}{
		"id":          uuid.New().String(),
		"kind":        "temperature sensor",
		"description": "kitchen thermometer",
	}

	for i := 0; i < 3; i++ {

		logrus.Infof("connecting to %s", u.String())

		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			logrus.Fatal("ws dial:", err)
		}
		defer c.Close()

		b, err := json.MarshalIndent(deviceInfo, "", "  ")
		if err != nil {
			logrus.Fatalf("marshal:", err)
			return
		}
		logrus.Infof("%s", b)

		err = c.WriteMessage(websocket.TextMessage, b)
		if err != nil {
			logrus.Info("write:", err)
			return
		}

		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		log.Printf("recv: %s", string(message))

		time.Sleep(200 * time.Millisecond)
		c.Close()
		time.Sleep(200 * time.Millisecond)
	}
}
