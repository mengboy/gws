package gws

import (
	"fmt"
	"github.com/gorilla/websocket"
	websocket2 "golang.org/x/net/websocket"
	"net/http"
	"testing"
)

func TestEngine_Run(t *testing.T) {
	e := New("/wshub", ":8080", func(writer http.ResponseWriter, request *http.Request) *websocket.Upgrader {
		return &websocket.Upgrader{
			HandshakeTimeout: 0,
			ReadBufferSize:   0,
			WriteBufferSize:  0,
			WriteBufferPool:  nil,
			Subprotocols:     nil,
			Error:            nil,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
			EnableCompression: false,
		}
	})
	if err := e.Run(); err != nil {
		panic(err)
	}
}

func Test_ConnectWS(t *testing.T)  {
	_, err := websocket2.Dial("ws://127.0.0.1:8080/wshub", "", "http://127.0.0.1:8080")
	if err != nil{
		t.Error(err)
	}
	fmt.Println(err)
}
