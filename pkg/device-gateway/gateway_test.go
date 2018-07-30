package gateway

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

var temperatureDeviceInfo = DeviceInfo{
	Kind:        "temperature sensor",
	Description: "kitchen thermometer",
}

type testDevice struct {
	DeviceInfo
	conn     *websocket.Conn
	response map[string]interface{}
}

func (td *testDevice) Close() {
	td.conn.Close()
}

func (td *testDevice) connect(u url.URL) error {
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("ws conn: %s", err)
	}
	td.conn = c

	if err := c.WriteJSON(td.DeviceInfo); err != nil {
		return fmt.Errorf("connect: %s", err)
	}

	var response map[string]interface{}
	if err := c.ReadJSON(&response); err != nil {
		return fmt.Errorf("reading response: %s", err)
	}

	td.response = response
	return nil
}

func TestConnect(t *testing.T) {
	resetTestState()
	srv := httptest.NewServer(http.HandlerFunc(connect))
	td := testDevice{
		temperatureDeviceInfo,
		nil,
		nil,
	}

	u := url.URL{
		Scheme: "ws",
		Host:   strings.Replace(srv.URL, "http://", "", -1),
		Path:   "/connect",
	}
	err := td.connect(u)
	if err != nil {
		t.Fatal(err)
	}
	defer td.Close()

	var expected interface{} = "OK"
	if td.response["response"] != expected {
		t.Fatalf("expected '%s', got '%s'", expected, td.response["response"])
	}
}

func TestConnectDisconnect(t *testing.T) {
	resetTestState()
	srv := httptest.NewServer(http.HandlerFunc(connect))
	td := testDevice{
		temperatureDeviceInfo,
		nil,
		nil,
	}

	u := url.URL{
		Scheme: "ws",
		Host:   strings.Replace(srv.URL, "http://", "", -1),
		Path:   "/connect",
	}
	err := td.connect(u)
	if err != nil {
		t.Fatal(err)
	}
	defer td.Close()

	registerdInfo := devices[td.ID]

	if registerdInfo.Connected != false {
		t.Fatalf("devices[%s][connected] != true, got %t", td.ID, registerdInfo.Connected)
	}

}

func resetTestState() {
	devices = map[string]DeviceInfo{}
}
