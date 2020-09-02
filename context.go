package websocket

import (
	"encoding/json"
	"io"
	"sync"
)

// Context 类型作为上下文数据缓存容器, 实现了几个常用的数据处理方法
type Context struct {
	reqBody []byte    // 缓存请求消息数据
	Writer  io.Writer // 响应接口
	written bool      // 是否已响应

	// 一个消息可能被多个处理器 handle, 如经常用前置处理器来验证授权信息, 记录日志等
	handlers []Handler              // 消息处理器列表
	index    int                    // 消息处理器索引
	values   map[string]interface{} // 缓存上下文临时数据
	Conn     *Conn                  // 服务器主动推送接口
}

var ctxPool = sync.Pool{
	New: func() interface{} {
		ctx := &Context{}
		ctx.init()
		return ctx
	},
}

func (ctx *Context) init() {
	ctx.reqBody = nil
	ctx.Writer = nil
	ctx.written = false
	ctx.handlers = nil
	ctx.index = 0
	ctx.values = map[string]interface{}{}
}

// handle 方法用于调用下一个处理器函数
func (ctx *Context) handle() {
	if ctx.index < len(ctx.handlers) {
		ctx.handlers[ctx.index].ServeWebsocket(ctx)
		ctx.index++
	}
}

// Next 方法将立即调用余下的处理器函数, 通常用在前置处理器中
func (ctx *Context) Next() {
	ctx.index++
	ctx.handle()
}

// Abort 用于终止所有剩余处理器的调用, 该方法不影响已调用的前置处理器
func (ctx *Context) Abort() {
	ctx.index = len(ctx.handlers)
}

// GetRequestBody 方法返回客户端请求消息数据
func (ctx *Context) GetRequestBody() []byte {
	return ctx.reqBody
}

// BindJSON 将请求消息数据绑定到 obj 中
func (ctx *Context) BindJSON(obj interface{}) error {
	return json.Unmarshal(ctx.reqBody, obj)
}

// SendData 将 data 数据并发送给客户端, 且终止剩余处理器的调用
func (ctx *Context) SendData(data []byte) error {
	if ctx.written {
		panic("multiple response data")
	}
	defer func() {
		ctx.written = true
		ctx.Abort()
	}()
	_, err := ctx.Writer.Write(data)
	return err
}

// SendJSON 将 obj 转转为 JSON 数据并发送给客户端, 且终止剩余处理器的调用
func (ctx *Context) SendJSON(obj interface{}) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return ctx.SendData(data)
}

func (ctx *Context) ClientIP() string {
	return ctx.Conn.conn.RemoteAddr().String()
}
