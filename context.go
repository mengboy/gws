package gws

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

type HeartBeatConf struct {
	RetryTimes       int // 重试次数
	HeartBeatTimeOut int
	HeartBeatChan    chan struct{}
	CurrentTimes     int
}

// context
type Context struct {
	ID      string // 连接id
	Conn    *websocket.Conn
	Writer  http.ResponseWriter
	Request *http.Request
	Logger  Log
	*HeartBeatConf
	StrMap       map[string]string
	Int64Map     map[string]int64
	Int32Map     map[string]int32
	InterfaceMap map[string]interface{}
	Timer        map[string]*time.Timer
	Group        *group // 所属组
	sync.Mutex
	*Engine
}

func (c *Context) SetInt64(key string, val int64) {
	c.Int64Map[key] = val
}

func (c *Context) GetInt64(key string) int64 {
	return c.Int64Map[key]
}

func (c *Context) SetInt32(key string, val int32) {
	c.Int32Map[key] = val
}

func (c *Context) GetInt32(key string) int32 {
	return c.Int32Map[key]
}

func (c *Context) SetInterface(key string, val interface{}) {
	c.InterfaceMap[key] = val
}

func (c *Context) GetInterface(key string) interface{} {
	return c.InterfaceMap[key]
}

// 心跳检测 连接是否die
func (c *Context) IfDie() bool {
	if c.RetryTimes == 0 && c.CurrentTimes == 0 {
		return true
	}
	return c.CurrentTimes >= c.RetryTimes
}

func (c *Context) IsOwner() bool {
	return c.Group.owner == c
}

// TODO 解析数据
func (c *Context) ParseData(msgByte []byte, ob interface{}) error {
	return json.Unmarshal(msgByte, ob)
}

// 设置参数
func (c *Context) SetString(key string, val string) {
	c.StrMap[key] = val
}

// 获取参数
func (c *Context) GetString(key string) string {
	return c.StrMap[key]
}

// 设置timer
func (c *Context) SetTimer(key string, timer *time.Timer) {
	if c.Timer == nil {
		c.Timer = map[string]*time.Timer{
			key: timer,
		}
		return
	}
	c.Timer[key] = timer
}

// stop timer
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

// 清除timer
func (c *Context) ClearTimer() {
	for _, t := range c.Timer {
		if t != nil {
			t.Stop()
		}
	}
	c.Timer = nil
}

// error msg
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

func (c *Context) Ping() {
	_ = c.SendText(Message{
		Path: "ping",
		Data: nil,
	})
}

func (c *Context) Pong() {
	_ = c.SendText(Message{
		Path: "pong",
		Data: nil,
	})
}

// TODO heart beat time
func (c *Context) StartHeartBeat() {
	go func() {
		c.Ping()
		for {
			ticker := time.NewTicker(10 * time.Second)
			select {
			case <-c.HeartBeatChan:
				c.Ping()
				time.Sleep(1 * time.Second)
			case <-ticker.C:
				c.CurrentTimes++
				if c.CurrentTimes < c.RetryTimes {
					c.Ping()
				}
			}
		}
	}()
}
