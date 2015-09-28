package ppu

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/jpeg"
	"net/http"

	"github.com/gorilla/websocket"
)

type webSocketRenderer struct {
	conn *websocket.Conn
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

	ws.conn = conn

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
	if ws.conn == nil {
		return
	}

	buf := bytes.NewBuffer(make([]byte, 0))

	err := jpeg.Encode(buf, img, &jpeg.Options{50})
	//err := png.Encode(buf, img)

	if err != nil {
		return
	}

	ws.conn.WriteMessage(websocket.TextMessage, []byte(base64.StdEncoding.EncodeToString(buf.Bytes())))
}
