package gws

type HandlerFunc func(ctx *Context, msg Msg)

type Router struct {
	Handler map[string]HandlerFunc
}

func (r *Router) AddRouter(router string, handle HandlerFunc) {
	r.Handler[router] = handle
}
