package main

import (
	"github.com/morgine/log"
	"github.com/morgine/websocket"
	"net/http"
)

func main() {
	var addr = ":4523"
	logR := websocket.Use(websocket.New1Logger())

	logR.HandleFunc(nil, websocket.NotFound)
	logR.HandleFunc("chat", func(ctx *websocket.Context) {

	})

	server := websocket.NewServer(routeGetter, nil, nil)

	websocket.PrintRoutes(websocket.DefaultRouter.Routes())

	http.Handle("/ws", server)

	log.Info.Printf("listen and serve: http://localhost%s", addr)

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Error.Fatal(err)
	}
}

type header struct {
	path string
}

func routeGetter(ctx *websocket.Context) (route interface{}) {
	h := &header{}
	ctx.BindJSON(h)
	return h.path
}
