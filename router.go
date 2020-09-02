package websocket

import (
	"reflect"
)

var DefaultRouter = NewRouter()

type Router interface {
	Handle(route interface{}, handler ...Handler)
	HandleFunc(route interface{}, handler ...HandlerFunc)
	Use(handler ...Handler) Router
	Match(route interface{}) []Handler
	Routes() map[interface{}][]Handler
}

func NewRouter() Router {
	return mux{}
}

type mux map[interface{}][]Handler

func (m mux) Routes() map[interface{}][]Handler {
	return m
}

func (m mux) HandleFunc(route interface{}, handler ...HandlerFunc) {
	m.Handle(route, handlersFToH(handler)...)
}

func (m mux) Match(route interface{}) []Handler {
	return m[route]
}

// 添加路由，如果 route 为 nil, 则表示匹配所有路径
func (m mux) Handle(route interface{}, handler ...Handler) {
	if route != nil && !reflect.TypeOf(route).Comparable() {
		panic("route is not comparable")
	}
	m[route] = append(m[route], handler...)
}

// 添加前置处理器
func (m mux) Use(handler ...Handler) Router {
	return &withHandlers{withs: handler, mux: m}
}

type withHandlers struct {
	withs []Handler
	mux   mux
}

func (w *withHandlers) Routes() map[interface{}][]Handler {
	return w.mux
}

// 匹配路由, 如果未匹配到路由, 则尝试用 nil 匹配
func (w *withHandlers) Match(route interface{}) []Handler {
	hs := w.mux.Match(route)
	if hs == nil && route != nil {
		return w.mux.Match(nil)
	}
	return hs
}

func (w *withHandlers) Handle(route interface{}, handler ...Handler) {
	w.mux.Handle(route, append(w.withs, handler...)...)
}

func (w *withHandlers) HandleFunc(route interface{}, handler ...HandlerFunc) {
	w.Handle(route, handlersFToH(handler)...)
}

func handlersFToH(handlers []HandlerFunc) []Handler {
	var hs = make([]Handler, len(handlers))
	for i, h := range handlers {
		hs[i] = h
	}
	return hs
}

func (w *withHandlers) Use(handler ...Handler) Router {
	return &withHandlers{
		withs: append(w.withs[:], handler...),
		mux:   w.mux,
	}
}
