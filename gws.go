package gws

import (
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"net/http"
)

type Engine struct {
	Path         string
	Port         string
	Upgrader     func(writer http.ResponseWriter, request *http.Request) websocket.Upgrader
	ResHeader    map[string]http.Header // 响应header
	Timeout      int64                  // websocket 持续时间
	router       *Router                // 路由
	MsgChanCap   int                    // 容量
	Logger       Log                    // log
	CreateConnID func() string          // 生成唯一id
	Hook
}

// 初始化engine
func New(path string, port string,
	upgrader func(writer http.ResponseWriter, request *http.Request) websocket.Upgrader,
	msgChanCap int, logger Log) *Engine {
	if logger == nil {
		logger = &Logger{
			Logger: zap.NewNop(),
		}
	}
	e := &Engine{
		Path:       path,
		Port:       port,
		Upgrader:   upgrader,
		router:     &Router{Handler: map[string]HandlerFunc{}},
		MsgChanCap: msgChanCap,
		Logger:     logger,
	}
	e.ProcessFunc = defaultProcess
	e.CreateConnID = UUID
	return e
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

func (e *Engine) AddRouter(router string, handle HandlerFunc) {
	e.router.AddRouter(router, handle)
}

func (e *Engine) GetRoute(router string) HandlerFunc {
	return e.router.Handler[router]
}

func (e *Engine) SetLogger(logger Log) {
	e.Logger = logger
}

func (e *Engine) Run() error {
	http.HandleFunc(e.Path, func(writer http.ResponseWriter, request *http.Request) {
		for _, f := range e.BeforeConn {
			f(writer, request)
		}
		conn, err := e.Upgrader(writer, request).Upgrade(writer, request, e.ResHeader[e.Path])
		if err != nil {
			return
		}
		ctx := &Context{
			ID:        e.CreateConnID(),
			Conn:      conn,
			Writer:    writer,
			Request:   request,
			MsgChan:   make(chan []byte, e.MsgChanCap),
			CloseChan: make(chan struct{}, 0),
			val:       map[string]string{},
			Logger:    e.Logger,
		}
		for _, f := range e.AfterConn {
			f(ctx)
		}
		e.ProcessFunc(ctx)
	})
	return http.ListenAndServe(e.Port, nil)
}

func defaultProcess(c *Context) {
	go func() {
	Loop:
		for {
			select {
			case msg := <-c.MsgChan:
				go func() {
					wm := &Message{}
					if err := c.ParseData(msg, wm); err != nil {
						c.Logger.Error(c, ErrorParseMsg+err.Error())
						return
					}
					handleFunc := c.Engine.GetRoute(wm.Router)
					if handleFunc == nil {
						c.Logger.Error(c, ErrorNotFountRouter+wm.Router)
						return
					}
					handleFunc(c, wm)
					return
				}()
			case <-c.CloseChan:
				break Loop
			}
		}
	}()

	for {
		msgType, msg, err := c.Conn.ReadMessage()
		if err != nil {
			c.Logger.Error(c, ErrorReadMsg+err.Error())
			// read msg failed process
			for _, f := range c.Engine.ReadErr {
				f(c, msgType, err)
			}
			c.CloseRead()
			break
		}
		// read msg succ
		for _, f := range c.Engine.AfterRead {
			f(c, msgType, msg)
		}
		c.MsgChan <- msg
	}
}
