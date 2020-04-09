package gws

import "net/http"

type Hook struct {
	BeforeConn  []func(writer http.ResponseWriter, request *http.Request) // 创建连接前插件
	AfterConn   []func(c *Context)                                        // 创建连接后插件
	ProcessFunc func(c *Context)                                          // 处理成功连接
	AfterRead   []func(c *Context, msgType int, msg []byte)               // 成功读取msg
	ReadErr     []func(c *Context, msgType int, err error)                // 读取msg error

	AfterJoinGroup  []func(c *Context, g *group) // 加入group
	AfterLeaveGroup []func(c *Context, g *group) // 离开group
}
