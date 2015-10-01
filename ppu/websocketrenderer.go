package ppu

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type webSocketRenderer struct {
	conns []*websocket.Conn
}

func NewWebSocketRenderer(endpoint string) Renderer {
	r := &webSocketRenderer{}

	http.HandleFunc(endpoint, r.handler)

	return r
}

func (ws *webSocketRenderer) handler(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	h := http.Header{}

	h["Sec-WebSocket-Protocol"] = []string{"nesrender"}
	conn, err := upgrader.Upgrade(w, r, h)

	if err != nil {
		panic("Unable to upgrade connection")
	}

	ws.conns = append(ws.conns, conn)

	for {
		t, _, err := conn.ReadMessage()
		if err != nil {
			break
		}

		if t == websocket.CloseMessage {
			break
		}
	}
}

func (ws *webSocketRenderer) Render(img image.Image) {
	if len(ws.conns) == 0 {
		return
	}

	buf := bytes.NewBuffer(make([]byte, 0))

	//err := jpeg.Encode(buf, img, &jpeg.Options{100})
	err := png.Encode(buf, img)

	if err != nil {
		return
	}

	base64img := []byte(base64.StdEncoding.EncodeToString(buf.Bytes()))

	var wg sync.WaitGroup
	wg.Add(len(ws.conns))
	for _, conn := range ws.conns {
		go func() {
			defer wg.Done()
			conn.WriteMessage(websocket.TextMessage, base64img)
		}()
	}

	wg.Wait()
}
