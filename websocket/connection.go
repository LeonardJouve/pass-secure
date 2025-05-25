package websocket

import (
	"encoding/json"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/LeonardJouve/pass-secure/database/queries"
	"github.com/gofiber/contrib/websocket"
)

type WebsocketConnection struct {
	SessionId    SessionId
	User         queries.User
	Connection   *websocket.Conn
	PongChannel  *PongChannel
	CloseChannel *CloseChannel
	sync.WaitGroup
	sync.Mutex
}

type WebsocketConnections struct {
	Connections map[SessionId]*WebsocketConnection
	sync.Mutex
}

func (websocketConnections *WebsocketConnections) add(websocketConnection *WebsocketConnection) {
	websocketConnections.Lock()
	defer websocketConnections.Unlock()

	websocketConnections.Connections[websocketConnection.SessionId] = websocketConnection
}

func (websocketConnections *WebsocketConnections) remove(websocketConnection *WebsocketConnection) {
	websocketConnections.Lock()
	defer websocketConnections.Unlock()

	delete(websocketConnections.Connections, websocketConnection.SessionId)
}

func (websocketConnections *WebsocketConnections) get(sessionId SessionId) (*WebsocketConnection, bool) {
	websocketConnections.Lock()
	defer websocketConnections.Unlock()

	websocketConnection, ok := websocketConnections.Connections[sessionId]

	return websocketConnection, ok
}

func (websocketConnection *WebsocketConnection) isInChannel(channel ChannelName) bool {
	websocketChannel, ok := websocketChannels.get(channel)
	if !ok {
		return false
	}

	if _, ok := websocketChannel[websocketConnection.SessionId]; !ok {
		return false
	}

	return true
}

func (websocketConnection *WebsocketConnection) isAllowedToJoinChannel(channel ChannelName) bool {
	switch {
	// TODO
	default:
		return false
	}
}

func (websocketConnection *WebsocketConnection) writeMessage(messageType MessageType, message MessageContent) bool {
	message["type"] = messageType

	marshaledMessage, err := json.Marshal(message)
	if err != nil {
		return false
	}

	websocketConnection.Lock()
	defer websocketConnection.Unlock()
	if err := websocketConnection.Connection.WriteMessage(websocket.TextMessage, marshaledMessage); err != nil {
		return false
	}

	return true
}

func (websocketConnection *WebsocketConnection) close() {
	select {
	case _, ok := <-*websocketConnection.CloseChannel:
		if ok {
			close(*websocketConnection.CloseChannel)
		}
	default:
	}
	websocketConnection.Connection.SetReadDeadline(time.Now())
	unregisterChannel <- websocketConnection
}

func (websocketConnection *WebsocketConnection) handlePingPong() {
	websocketConnection.Add(1)
	defer websocketConnection.Done()

	websocketTimeoutString := os.Getenv("WEBSOCKET_TIMEOUT_IN_SECOND")
	websocketTimeout, err := strconv.ParseInt(websocketTimeoutString, 10, 64)
	if err != nil {
		websocketConnection.close()
		return
	}
	timeout := time.Duration(websocketTimeout) * time.Second

	pingTicker := time.NewTicker(2 * timeout)
	defer pingTicker.Stop()

	timeoutTicker := time.NewTicker(timeout)
	timeoutTicker.Stop()
	defer timeoutTicker.Stop()

	hasPong := true

	for {
		select {
		case <-*websocketConnection.CloseChannel:
			return
		case <-*websocketConnection.PongChannel:
			hasPong = true
			timeoutTicker.Stop()
		case <-timeoutTicker.C:
			websocketConnection.close()
			return
		case <-pingTicker.C:
			if !hasPong {
				continue
			}
			websocketConnection.writeMessage(PING_TYPE, MessageContent{})
			timeoutTicker.Reset(timeout)
			hasPong = false
		}
	}
}

func (websocketConnections *WebsocketConnections) writeGlobalMessage(messageType MessageType, message MessageContent) {
	// TODO: mutex ?
	for _, websocketConnection := range websocketConnections.Connections {
		websocketConnection.writeMessage(messageType, message)
	}
}

func (websocketConnections *WebsocketConnections) writeChannelMessage(channel ChannelName, messageType MessageType, message MessageContent) {
	message["channel"] = channel

	// TODO: remove global variable dependency
	websocketChannel, ok := websocketChannels.get(channel)
	if !ok {
		return
	}

	for sessionId := range websocketChannel {
		websocketConnection, ok := websocketConnections.get(sessionId)
		if !ok {
			continue
		}

		websocketConnection.writeMessage(messageType, message)
	}
}
