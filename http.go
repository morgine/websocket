package websocket

import "net/http"

// RouteGetter 类型用于定义从客户端消息中获取路由地址的方法
type RouteGetter func(ctx *Context) (route interface{}, err error)

func NewHTTPHandler(getter RouteGetter) http.Handler {

}
