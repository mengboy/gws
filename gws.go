package gws

import (
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"net/http"
)

type Engine struct {
	Port          string
	Upgrader      func(writer http.ResponseWriter, request *http.Request) *websocket.Upgrader
	ResHeader     map[string]http.Header // 响应header
	OpenHeartBeat bool
	Timeout       int64         // websocket 持续时间
	router        *Router       // 路由
	Logger        Log           // log
	CreateConnID  func() string // 生成唯一id
	*HeartBeatConf
	*Hook
}

// 初始化engine
func New(port string,
	upgrader func(writer http.ResponseWriter, request *http.Request) *websocket.Upgrader, heartBeat bool, heartBeatConf *HeartBeatConf) *Engine {
	// 初始化默认值
	if heartBeat {
		if heartBeatConf == nil {
			heartBeatConf = &HeartBeatConf{
				RetryTimes:       0,
				HeartBeatTimeOut: 0,
				HeartBeatChan:    make(chan struct{}, 0),
				CurrentTimes:     0,
			}
		}
		if heartBeatConf.HeartBeatChan == nil {
			heartBeatConf.HeartBeatChan = make(chan struct{}, 0)
		}
	}
	e := &Engine{
		Port:     port,
		Upgrader: upgrader,
		router:   &Router{Handler: map[string]HandlerFunc{}},
		Logger: &Logger{
			Logger: zap.NewNop(),
		},
		OpenHeartBeat: heartBeat,
		HeartBeatConf: heartBeatConf,
		Hook:          &Hook{},
		ResHeader:     map[string]http.Header{},
	}
	if heartBeat {
		e.AddHeartBeatHandler()
	}
	e.ProcessFunc = defaultProcess
	e.CreateConnID = UUID
	return e
}

// 开启心跳
func (e *Engine) AddHeartBeatHandler(f ...func(ctx *Context, msg Msg)) {
	// 处理ping请求
	e.AddPath("ping", func(ctx *Context, msg Msg) {
		ctx.Pong()
	})
	// 处理pong 响应
	e.AddPath("pong", func(ctx *Context, msg Msg) {
		ctx.HeartBeatChan <- struct{}{}
	})
	if f != nil {
		e.AddPath("ping", f[0])
	}
}

// 创建连接前
func (e *Engine) AddBefore(beforeConn ...func(writer http.ResponseWriter, request *http.Request)) {
	e.BeforeConn = append(e.BeforeConn, beforeConn...)
}

// 生成id
func (e *Engine) SetCreateConnIDFunc(createID func() string) {
	e.CreateConnID = createID
}

// 设置处理 msg 函数
func (e *Engine) SetProcessFunc(processFunc func(c *Context)) {
	e.ProcessFunc = processFunc
}

func (e *Engine) AddAfterRead(afterRead ...func(c *Context, msgType int, msg []byte)) {
	e.AfterRead = append(e.AfterRead, afterRead...)
}

func (e *Engine) AddAfterConn(afterConn ...func(c *Context)) {
	e.AfterConn = append(e.AfterConn, afterConn...)
}

func (e *Engine) SetResHeader(path string, header http.Header) {
	if e.ResHeader == nil {
		e.ResHeader = map[string]http.Header{}
	}
	e.ResHeader[path] = header
}

func (e *Engine) AddPath(path string, handle HandlerFunc) {
	e.router.AddPath(path, handle)
}

func (e *Engine) GetHandlerFunc(path string) HandlerFunc {
	return e.router.Handler[path]
}

func (e *Engine) SetLogger(logger Log) {
	e.Logger = logger
}


// 添加ws handler
func (e *Engine) AddWsHandler(path string, handlerFunc http.HandlerFunc) {
	if handlerFunc == nil {
		// default
		handlerFunc = func(writer http.ResponseWriter, request *http.Request) {
			e.Hook.StartBeforeConn(writer, request)
			conn, err := e.Upgrader(writer, request).Upgrade(writer, request, e.ResHeader[path])
			if err != nil {
				return
			}
			ctx := &Context{
				ID:            e.CreateConnID(),
				Conn:          conn,
				Writer:        writer,
				Request:       request,
				val:           map[string]string{},
				Logger:        e.Logger,
				Engine:        e,
				HeartBeatConf: e.HeartBeatConf,
			}
			if e.OpenHeartBeat {
				ctx.StartHeartBeat()
			}
			e.Hook.StartAfterConn(ctx)
			e.Hook.StartProcessFunc(ctx)
		}
	}
	http.HandleFunc(path, handlerFunc)
}

func (e *Engine) Run() error {
	e.Logger.Info(nil, "start server")
	return http.ListenAndServe(e.Port, nil)
}

func defaultProcess(c *Context) {
	msgChan := make(chan []byte, 100)
	closeChan := make(chan struct{})
	go func() {
	Loop:
		for {
			select {
			case msg := <-msgChan:
				go func() {
					wm := &Message{}
					if err := c.ParseData(msg, wm); err != nil {
						c.Logger.Error(c, ErrorParseMsg+err.Error())
						return
					}
					handleFunc := c.Engine.GetHandlerFunc(wm.Path)
					if handleFunc == nil {
						c.Logger.Error(c, ErrorNotFountRouterMsg+wm.Path)
						return
					}
					// TODO 处理
					handleFunc(c, wm)
					return
				}()
			case <-closeChan:
				break Loop
			}
		}
	}()

	for {
		msgType, msg, err := c.Conn.ReadMessage()
		if err != nil {
			c.Logger.Error(c, ErrorReadMsg+err.Error())
			// read msg failed process
			c.Engine.Hook.StartReadErrFunc(c, msgType, err)
			closeChan <- struct{}{}
			break
		}
		// read msg succ
		c.Engine.Hook.StartAfterRead(c, msgType, msg)
		msgChan <- msg
	}
}
