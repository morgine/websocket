package websocket

import (
	"github.com/morgine/log"
	"reflect"
	"runtime"
)

func PrintRoutes(routes map[interface{}][]Handler) {
	for route, handlers := range routes {
		if lst := len(handlers) - 1; lst >= 0 {
			lstH := handlers[lst]
			hName := getHandlerName(lstH)
			log.Init.Printf("[WS] %-10v --> %s (%d handlers)\n", route, hName, lst+1)
		}
	}
}

func getHandlerName(h Handler) string {
	rv := reflect.ValueOf(h)
	rt := rv.Type()
	switch k := rt.Kind(); k {
	case reflect.Func:
		return runtime.FuncForPC(rv.Pointer()).Name()
	default:
		for {
			if k == reflect.Ptr || k == reflect.Interface {
				rt = rt.Elem()
				break
			}
		}
		return rt.PkgPath() + "." + rt.Name()
	}
}
