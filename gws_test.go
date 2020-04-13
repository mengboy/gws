package gws

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	websocket2 "golang.org/x/net/websocket"
	"net/http"
	"testing"
	"time"
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
	e.AddPath("ping", func(ctx *Context, msg Msg) {
		_ = ctx.SendText(Message{
			Path: "pong",
			Data: nil,
		})
	})

	go func() {
		if err := e.Run(); err != nil {
			panic(err)
		}

	}()
	time.Sleep(1 * time.Second)
	conn, err := websocket2.Dial("ws://127.0.0.1:8080/wshub", "", "http://127.0.0.1:8080")
	if err != nil {
		t.Error(err)
	}
	msg := Message{
		Path: "ping",
		Data: nil,
	}
	jsonByte, _ := json.Marshal(msg)
	_, err = conn.Write(jsonByte)
	if err != nil {
		t.Error(err)
	}
	var res = make([]byte, 2048)
	n, err := conn.Read(res)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, string(res[0:n]), "{\"path\":\"pong\",\"data\":null}")
}
