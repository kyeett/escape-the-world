package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/connect"}
	logrus.Infof("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		logrus.Fatal("ws dial:", err)
	}
	defer c.Close()

	deviceInfo := map[string]interface{}{
		"id":          uuid.New().String(),
		"kind":        "temperature sensor",
		"description": "kitchen thermometer",
	}

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

	/*
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:

				if err != nil {
					log.Println("write:", err)
					return
				}
			case <-interrupt:
				log.Println("interrupt")

				// Cleanly close the connection by sending a close message and then
				// waiting (with timeout) for the server to close the connection.
				err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					log.Println("write close:", err)
					return
				}
				select {
				case <-done:
				case <-time.After(time.Second):
				}
				return
			}
		}*/
}
