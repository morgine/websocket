package websocket

import (
	"reflect"
)

var DefaultRouter = NewRouter()

type Router interface {
	Handle(route interface{}, handler ...Handler)
	Use(handler ...Handler) Router
	Match(route interface{}) []Handler
}

func NewRouter() Router {
	return mux{}
}

type mux map[interface{}][]Handler

func (m mux) Match(route interface{}) []Handler {
	return m[route]
}

func (m mux) Handle(route interface{}, handler ...Handler) {
	if !reflect.TypeOf(route).Comparable() {
		panic("route is not comparable")
	}
	m[route] = append(m[route], handler...)
}

func (m mux) Use(handler ...Handler) Router {
	return &withHandlers{withs: handler, mux: m}
}

type withHandlers struct {
	withs []Handler
	mux   mux
}

func (w *withHandlers) Match(route interface{}) []Handler {
	return w.mux.Match(route)
}

func (w *withHandlers) Handle(route interface{}, handler ...Handler) {
	w.mux.Handle(route, append(w.withs, handler...)...)
}

func (w *withHandlers) Use(handler ...Handler) Router {
	return &withHandlers{
		withs: append(w.withs[:], handler...),
		mux:   w.mux,
	}
}
