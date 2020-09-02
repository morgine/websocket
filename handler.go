package websocket

// websocket 消息处理接口, 该接口规定义了 websocket 消息处理方法
type Handler interface {
	ServeWebsocket(ctx *Context)
}

// HandlerFunc 类型实现了 Handler 接口, 方便实现该接口
type HandlerFunc func(ctx *Context)

func (h HandlerFunc) ServeWebsocket(ctx *Context) {
	h(ctx)
}

func HandleFunc(route interface{}, handler ...Handler) {
	DefaultRouter.Handle(route, handler...)
}

func Use(handler ...Handler) Router {
	return DefaultRouter.Use(handler...)
}

func NotFound(ctx *Context) {
	ctx.SendData([]byte("route not found"))
}

func NotFoundHandler() Handler {
	return HandlerFunc(NotFound)
}
