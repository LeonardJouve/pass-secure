package websocket

import (
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/google/uuid"
)

type UserConnections = []*WebsocketConnection
type WriteChannel = chan WriteWork

type WebsocketConnection struct {
	id                     uuid.UUID
	userId                 int64
	connection             *websocket.Conn
	closeChannel           CloseChannel
	closeGracefullyChannel CloseChannel
	closeGracefullyOnce    sync.Once
	writeTimeout           time.Duration
	sync.WaitGroup
	sync.Mutex
	sync.Once
}

type WebsocketConnections struct {
	connections  map[int64]UserConnections
	writeChannel WriteChannel
	closeChannel CloseChannel
	sync.WaitGroup
	sync.Mutex
	sync.Once
}

type WriteWork struct {
	connection *WebsocketConnection
	content    []byte
}

const CLOSE_GRACEFULLY_TIMEOUT = 5 * time.Second

func (w *WebsocketConnection) writeBytes(content []byte) {
	w.Lock()
	defer w.Unlock()

	w.connection.SetWriteDeadline(time.Now().Add(w.writeTimeout))
	w.connection.WriteMessage(websocket.TextMessage, content)
}

func (w *WebsocketConnection) ping() {
	w.Lock()
	defer w.Unlock()

	w.connection.SetWriteDeadline(time.Now().Add(w.writeTimeout))
	w.connection.WriteMessage(websocket.PingMessage, []byte{})
}

func (w *WebsocketConnection) askForClosure() {
	w.connection.SetWriteDeadline(time.Now().Add(w.writeTimeout))
	w.connection.WriteMessage(websocket.CloseMessage, []byte{})

	select {
	case <-w.closeGracefullyChannel:
	case <-time.After(CLOSE_GRACEFULLY_TIMEOUT):
		w.closeGracefullyOnce.Do(func() {
			close(w.closeGracefullyChannel)
		})
	}
}

func (w *WebsocketConnection) closeGracefully() {
	w.closeGracefullyOnce.Do(func() {
		close(w.closeGracefullyChannel)
		w.close()
	})
}

func (w *WebsocketConnection) close() {
	w.Do(func() {
		w.askForClosure()
		close(w.closeChannel)
		w.Wait()
		w.connection.Close()
	})
}

func (w *WebsocketConnection) handlePingPong(timeout time.Duration) {
	defer w.Done()

	w.connection.SetReadDeadline(time.Now().Add(timeout))
	w.connection.SetPongHandler(func(string) error {
		w.connection.SetReadDeadline(time.Now().Add(timeout))
		return nil
	})

	pingTicker := time.NewTicker(timeout * 9 / 10)
	defer pingTicker.Stop()

	for {
		select {
		case <-w.closeChannel:
			return
		case <-pingTicker.C:
			w.ping()
		}
	}
}

func (w *WebsocketConnection) readMessages() {
	defer w.Done()

	w.Add(1)
	go func() {
		defer w.Done()

		for {
			websocketMessageType, _, err := w.connection.ReadMessage()
			if err != nil {
				return
			}

			if websocketMessageType == websocket.CloseMessage {
				w.closeGracefully()
				return
			}
		}
	}()

	<-w.closeChannel
	w.connection.SetReadDeadline(time.Now())
}

func (w *WebsocketConnections) add(websocketConnection *WebsocketConnection) {
	w.Lock()
	defer w.Unlock()

	userConnections, ok := w.connections[websocketConnection.userId]
	if !ok {
		userConnections = []*WebsocketConnection{}
	}

	w.connections[websocketConnection.userId] = append(userConnections, websocketConnection)
}

func (w *WebsocketConnections) remove(websocketConnection *WebsocketConnection) {
	w.Lock()
	defer w.Unlock()

	userConnections, ok := w.connections[websocketConnection.userId]
	if !ok {
		return
	}

	for i, connection := range userConnections {
		if connection.id == websocketConnection.id {
			if len(w.connections[websocketConnection.userId]) == 1 {
				delete(w.connections, websocketConnection.userId)
			} else {
				w.connections[websocketConnection.userId] = append(userConnections[:i], userConnections[i+1:]...)
			}

			return
		}
	}
}

func (w *WebsocketConnections) broadcastNotification(notification Notification) {
	w.Lock()
	defer w.Unlock()

	for _, userConnections := range w.connections {
		for _, websocketConnection := range userConnections {
			w.writeChannel <- WriteWork{
				connection: websocketConnection,
				content:    []byte(notification.Message),
			}
		}
	}
}

func (w *WebsocketConnections) sendNotification(notification Notification) {
	w.Lock()
	defer w.Unlock()

	if notification.Broadcast {
		w.broadcastNotification(notification)
		return
	}

	for _, userId := range notification.UserIds {
		userConnections, ok := w.connections[userId]
		if !ok {
			continue
		}

		for _, websocketConnection := range userConnections {
			w.writeChannel <- WriteWork{
				connection: websocketConnection,
				content:    []byte(notification.Message),
			}
		}
	}
}

func (w *WebsocketConnections) handleWriteWorkers() {
	defer w.Done()

	for {
		select {
		case work := <-w.writeChannel:
			w.Add(1)
			go func() {
				defer w.Done()

				work.connection.writeBytes(work.content)
			}()
		case <-w.closeChannel:
			return
		}
	}
}

func (w *WebsocketConnections) close() {
	w.Do(func() {
		close(w.closeChannel)
		w.Wait()
		close(w.writeChannel)
	})
}
