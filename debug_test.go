package websocket

import "testing"

type structHandler struct {
}

func (s *structHandler) ServeWebsocket(ctx *Context) {
	panic("implement me")
}

func TestGetHandlerName(t *testing.T) {
	type testCase struct {
		got  string
		need string
	}
	var cases = []testCase{
		{getHandlerName(&structHandler{}), "github.com/morgine/websocket.structHandler"},
		{getHandlerName(NotFoundHandler()), "github.com/morgine/websocket.NotFound"},
	}
	for _, c := range cases {
		if c.got != c.need {
			t.Errorf("need: %s, got: %s\n", c.need, c.got)
		}
	}
}
