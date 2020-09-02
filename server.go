package websocket

import (
	"bytes"
	"github.com/gorilla/websocket"
	"github.com/morgine/log"
	"net/http"
	"runtime/debug"
	"time"
)

const (
	//發送隊列長度
	writeChanSize = 5

	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 1024
)

// RouteGetter 类型用于定义从客户端消息中获取路由地址的方法
type RouteGetter func(ctx *Context) (route interface{})

func NewServer(getter RouteGetter, router Router, upgrader *websocket.Upgrader) *Server {
	if router == nil {
		router = DefaultRouter
	}
	if upgrader == nil {
		upgrader = &websocket.Upgrader{}
	}
	return &Server{
		routeGetter: getter,
		router:      router,
		upgrader:    upgrader,
	}
}

type Server struct {
	routeGetter RouteGetter
	router      Router
	upgrader    *websocket.Upgrader
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	conn, err := s.upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Warning.Printf("gitbub.com/morgine/websocket.ServeHTTP upgrade websocket got error: %s\n", err)
	} else {
		c := NewConn(conn, func(ctx *Context) []Handler {
			// 从 Context 中获取路由地址, 该值可以是 nil
			route := s.routeGetter(ctx)
			return s.router.Match(route)
		})
		c.Run()
	}
}

type Conn struct {
	conn            *websocket.Conn
	handlerProvider func(ctx *Context) []Handler
	send            chan []byte
}

func NewConn(c *websocket.Conn, handlerProvider func(ctx *Context) []Handler) *Conn {
	return &Conn{conn: c, handlerProvider: handlerProvider}
}

func (c *Conn) SendMessage(data []byte) {
	c.send <- data
}

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// 开始监听消息
func (c *Conn) runReader() {
	defer func() {
		c.Close()
		if err := recover(); err != nil {
			log.Panic.Printf("websocket.runReader panics: %s\n", debug.Stack())
		}
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Warning.Printf("websocket.runReader() got close error: %s\n", err)
				return
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		response := c.handleMessage(message)
		c.SendMessage(response)
	}
}

// handleMessage 初始化 Context 并调用处理函数, 如果消息数据中没有路由信息或未匹配到路由, 则调用 NotFoundHandler 处理
func (c *Conn) handleMessage(message []byte) (response []byte) {
	ctx := ctxPool.Get().(*Context)
	defer func() {
		ctx.init()
		ctxPool.Put(ctx)
		if err := recover(); err != nil {
			log.Panic.Printf("websocket.handleMessage panics: %s\n", debug.Stack())
		}
	}()
	ctx.reqBody = message
	buff := bytes.NewBuffer(nil)
	ctx.Writer = buff
	handlers := c.handlerProvider(ctx)
	if len(handlers) == 0 {
		handlers = []Handler{NotFoundHandler()}
	}
	ctx.handlers = handlers
	ctx.Conn = c
	ctx.handle()
	return buff.Bytes()
}

//初始化鏈接接受隊列跟發送隊列
func (c *Conn) Run() {
	//發送隊列初始化
	c.send = make(chan []byte, writeChanSize)

	//开启发送队列
	go c.runReader()
	//开启接收队列
	go c.runWriter()
}

func (c *Conn) Close() {
	c.conn.Close()
	if c.send != nil {
		close(c.send)
	}
}

//像客戶端發送網絡包
func (c *Conn) runWriter() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
		if err := recover(); err != nil {
			log.Panic.Printf("websocket.runWriter panics: %s\n", debug.Stack())
		}
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
