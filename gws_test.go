package gws

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	websocket2 "golang.org/x/net/websocket"
	"net/http"
	"testing"
	"time"
)

func TestEngine_Run(t *testing.T) {
	e := New(":8080", func(writer http.ResponseWriter, request *http.Request) *websocket.Upgrader {
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
	}, false, nil)
	e.AddPath("/wshub", nil)
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

func Test_Heartbeat(t *testing.T) {
	e := New(":8080", func(writer http.ResponseWriter, request *http.Request) *websocket.Upgrader {
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
	}, true, nil)
	e.AddPath("/wshub", nil)
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
	if conn == nil {
		t.Error(errors.New("get conn failed"))
	}
	time.Sleep(1 * time.Second)
	errChan := make(chan error, 0)
	msgChan := make(chan []byte, 0)
	go readMsg(conn, errChan, msgChan)
	i := 0
LOOP:
	for {
		select {
		case err := <-errChan:
			t.Error(err)
			break LOOP
		case msg := <-msgChan:
			expectMsg := "{\"path\":\"ping\",\"data\":null}"
			if string(msg) != expectMsg {
				t.Error(fmt.Errorf("expect: %s, got: %s", expectMsg, msg))
			}
			i++
			if i == 10 {
				t.Log("succ")
				break LOOP
			}
		}
	}
}

func readMsg(wsConn *websocket2.Conn, errChan chan error, msgChan chan []byte) {
	for {
		var res = make([]byte, 2048)
		n, err := wsConn.Read(res)
		if err != nil {
			errChan <- err
		}
		msg := Message{}
		if err := json.Unmarshal(res[0:n], &msg); err != nil {
			errChan <- err
		}
		if msg.Path == "ping" {
			msg := Message{
				Path: "pong",
				Data: nil,
			}
			jsonByte, _ := json.Marshal(msg)
			_, err = wsConn.Write(jsonByte)
			if err != nil {
				errChan <- err
			}
			msgChan <- res[0:n]
		}
	}
}
