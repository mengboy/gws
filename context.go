package gws

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

// context
type Context struct {
	ID      string // 连接id
	Conn    *websocket.Conn
	Writer  http.ResponseWriter
	Request *http.Request
	Logger  Log
	val     map[string]string
	Timer   map[string]*time.Timer
	Group   *group // 所属组
	sync.Mutex
	*Engine
}

func (c *Context) ParseData(msgByte []byte, ob interface{}) error {
	return json.Unmarshal(msgByte, ob)
}

func (c *Context) Set(key string, val string) {
	c.val[key] = val
}

func (c *Context) Get(key string) string {
	return c.val[key]
}

func (c *Context) SetTimer(key string, timer *time.Timer) {
	if c.Timer == nil {
		c.Timer = map[string]*time.Timer{
			key: timer,
		}
		return
	}
	c.Timer[key] = timer
}

func (c *Context) CloseTimer(key string) {
	if c.Timer == nil {
		return
	}
	if t, ok := c.Timer[key]; ok {
		if t != nil {
			t.Stop()
		}
		delete(c.Timer, key)
	}
}

func (c *Context) ClearTimer() {
	for _, t := range c.Timer {
		if t != nil {
			t.Stop()
		}
	}
	c.Timer = nil
}

//error msg
// read: connection reset by peer
// use of closed network connection
// write: broken pipe
// websocket: close sent
//
func (c *Context) SendText(data interface{}) error {
	jsonMsg, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if c.Conn == nil {
		return errors.New(ErrorNotConnMsg)
	}

	c.Lock()
	defer c.Unlock()
	return c.Conn.WriteMessage(websocket.TextMessage, jsonMsg)
}

func (c *Context) SendJson(data interface{}) error {
	if c.Conn == nil {
		return errors.New(ErrorNotConnMsg)
	}
	c.Lock()
	defer c.Unlock()
	return c.Conn.WriteJSON(data)
}
