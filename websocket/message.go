package websocket

import "encoding/json"

type MessageType = string
type MessageContent = map[string]interface{}

type Message struct {
	MessageType MessageType    `json:"type"`
	Content     MessageContent `json:"content"`
}

func (m *Message) marshal() ([]byte, error) {
	return json.Marshal(m)
}
