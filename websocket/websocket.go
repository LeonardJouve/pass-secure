package websocket

type PongChannel = chan struct{}
type DatabaseNotificationChannel = chan string
type MessageChannel = chan *Message

var websocketConnections = WebsocketConnections{
	Connections: make(map[SessionId]*WebsocketConnection),
}
