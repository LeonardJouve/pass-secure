package websocket

import (
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
)

type PongChannel = chan struct{}
type UserConnections = []*WebsocketConnection

type WebsocketConnection struct {
	userId       int64
	connection   *websocket.Conn
	closeChannel CloseChannel
	pongChannel  PongChannel
	sync.WaitGroup
	sync.Mutex
}

type WebsocketConnections struct {
	connections map[int64]UserConnections
	sync.Mutex
}

func (w *WebsocketConnection) writeMessage(message Message) {
	content, err := message.marshal()
	if err != nil {
		return
	}

	w.writeBytes(content)
}

func (w *WebsocketConnection) writeBytes(content []byte) {
	w.Lock()
	defer w.Unlock()

	w.connection.WriteMessage(websocket.TextMessage, content)
}

func (w *WebsocketConnection) ping() {
	w.Lock()
	defer w.Unlock()

	w.connection.WriteMessage(websocket.PingMessage, []byte{})
}

func (w *WebsocketConnection) close() {
	// TODO mutex ?
	select {
	case _, ok := <-w.CloseChannel:
		if ok {
			// TODO
			close(w.CloseChannel)
		}
	default:
	}
	w.connection.SetReadDeadline(time.Now())

	// TODO: send close message and wait for close response
	w.connection.Close()
}

func (w *WebsocketConnection) handlePingPong(timeout time.Duration) {
	defer w.Done()

	pingTicker := time.NewTicker(2 * timeout)
	defer pingTicker.Stop()

	timeoutTicker := time.NewTicker(timeout)
	timeoutTicker.Stop()
	defer timeoutTicker.Stop()

	hasPong := true

	for {
		select {
		case <-w.closeChannel:
			return
		case <-w.pongChannel:
			hasPong = true
			timeoutTicker.Stop()
		case <-timeoutTicker.C:
			w.close()
			return
		case <-pingTicker.C:
			// TODO
			if !hasPong {
				continue
			}
			w.ping()
			timeoutTicker.Reset(timeout)
			hasPong = false
		}
	}
}

func (w *WebsocketConnections) add(websocketConnection *WebsocketConnection) {
	w.Lock()
	defer w.Unlock()

	if userConnections, ok := w.connections[websocketConnection.userId]; ok {
		w.connections[websocketConnection.userId] = append(userConnections, websocketConnection)
	} else {
		w.connections[websocketConnection.userId] = []*WebsocketConnection{websocketConnection}
	}
}

func (w *WebsocketConnections) remove(websocketConnection *WebsocketConnection) {
	// TODO
	w.Lock()
	defer w.Unlock()

	userConnections, ok := w.connections[websocketConnection.userId]
	if !ok {
		return
	}

	delete(w.connections, websocketConnection.SessionId)
}

func (w *WebsocketConnections) writeGlobalMessage(message Message) {
	content, err := message.marshal()
	if err != nil {
		return
	}

	w.Lock()
	defer w.Unlock()
	for _, userConnections := range w.connections {
		for _, websocketConnection := range userConnections {
			// TODO: use go routine with buffered channel
			websocketConnection.writeBytes(content)
		}
	}
}
