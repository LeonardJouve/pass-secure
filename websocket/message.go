package websocket

type MessageType = string
type MessageContent = map[string]interface{}

type Message struct {
	MessageType         MessageType
	Message             MessageContent
	WebsocketConnection *WebsocketConnection
}

// TODO: Marshal / Unmarshal
// TODO: see websocket WriteJSON
