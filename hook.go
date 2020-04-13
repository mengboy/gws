package gws

import "net/http"

type Hook struct {
	BeforeConn      []func(writer http.ResponseWriter, request *http.Request) // 创建连接前插件
	AfterConn       []func(c *Context)                                        // 创建连接后插件
	ProcessFunc     func(c *Context)                                          // 处理成功连接
	AfterRead       []func(c *Context, msgType int, msg []byte)               // 成功读取msg
	ReadErr         []func(c *Context, msgType int, err error)                // 读取msg error
	CreateGroup     []func(g *group)                                          // 创建group成功
	AfterJoinGroup  []func(c *Context, g *group)                              // 加入group
	AfterLeaveGroup []func(c *Context, g *group)                              // 离开group
	GroupSendFailed []func(c *Context, g *group)                              // 处理群组发送消息失败的情况, 向c发送消息失败
}

func (h *Hook) StartBeforeConn(writer http.ResponseWriter, request *http.Request) {
	for _, f := range h.BeforeConn {
		f(writer, request)
	}
}

func (h *Hook) StartAfterConn(c *Context) {
	for _, f := range h.AfterConn {
		f(c)
	}
}

func (h *Hook) StartProcessFunc(c *Context) {
	h.ProcessFunc(c)
}

func (h *Hook) StartAfterRead(c *Context, msgType int, msg []byte) {
	for _, f := range h.AfterRead {
		f(c, msgType, msg)
	}
}

func (h *Hook) StartReadErrFunc(c *Context, msgType int, err error) {
	for _, f := range c.Engine.ReadErr {
		f(c, msgType, err)
	}
}

func (h *Hook) StartCreateGroup(g *group) {
	for _, f := range h.CreateGroup {
		f(g)
	}
}

func (h *Hook) StartAfterJoinGroup(c *Context, g *group) {
	for _, f := range g.gm.AfterJoinGroup {
		f(c, g)
	}
}

func (h *Hook) StartAfterLeaveGroup(c *Context, g *group) {
	for _, f := range g.gm.AfterLeaveGroup {
		f(c, g)
	}
}

func (h *Hook) StartGroupSendFailed(c *Context, g *group) {
	for _, f := range g.gm.GroupSendFailed {
		f(c, g)
	}
}
