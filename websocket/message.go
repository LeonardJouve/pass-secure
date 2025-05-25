package websocket

type MessageType = string
type MessageContent = map[string]interface{}

type Message struct {
	Channel             ChannelName
	MessageType         MessageType
	Message             MessageContent
	WebsocketConnection *WebsocketConnection
}

// TODO: Marshal
