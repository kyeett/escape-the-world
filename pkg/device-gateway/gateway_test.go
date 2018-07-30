package gateway

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func TestConnect(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(connect))
	u := url.URL{
		Scheme: "ws",
		Host:   strings.Replace(srv.URL, "http://", "", -1),
		Path:   "/connect",
	}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatal("ws conn:", err)
	}
	defer c.Close()

	deviceInfo := map[string]interface{}{
		"id":          uuid.New().String(),
		"kind":        "temperature sensor",
		"description": "kitchen thermometer",
	}

	if err := c.WriteJSON(deviceInfo); err != nil {
		t.Fatal("ws write:", err)
	}

	var response map[string]interface{}
	if err := c.ReadJSON(&response); err != nil {
		t.Fatal("ws read:", err)
	}

	var expected interface{} = "OK"
	if response["response"] != expected {
		t.Fatalf("expected '%s', got '%s'", expected, response["response"])
	}

}
