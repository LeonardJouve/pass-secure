package websocket

import "sync"

type ChannelName = string
type WebsocketChannel = map[SessionId]struct{}

type WebsocketChannels struct {
	Channels map[ChannelName]WebsocketChannel
	sync.Mutex
}

func (websocketChannels *WebsocketChannels) add(websocketConnection *WebsocketConnection, channel ChannelName) {
	websocketChannels.Lock()
	defer websocketChannels.Unlock()

	websocketChannels.Channels[channel][websocketConnection.SessionId] = struct{}{}
}

func (websocketChannels *WebsocketChannels) remove(websocketConnection *WebsocketConnection, channel ChannelName) {
	websocketChannels.Lock()
	defer websocketChannels.Unlock()

	delete(websocketChannels.Channels[channel], websocketConnection.SessionId)
}

func (websocketChannels *WebsocketChannels) get(channel ChannelName) (WebsocketChannel, bool) {
	websocketChannels.Lock()
	defer websocketChannels.Unlock()

	websocketChannel, ok := websocketChannels.Channels[channel]

	return websocketChannel, ok
}
