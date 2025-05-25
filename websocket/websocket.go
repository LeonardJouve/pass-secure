package websocket

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/LeonardJouve/pass-secure/database"
	"github.com/LeonardJouve/pass-secure/database/queries"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

// TODO: use websocket.* for PING PONG message types

type SessionId = string
type PongChannel = chan struct{}
type CloseChannel = chan struct{}

const (
	JOIN_TYPE            = "join"
	LEAVE_TYPE           = "leave"
	REGISTER_TYPE        = "register"
	UNREGISTER_TYPE      = "unregister"
	PING_TYPE            = "ping"
	PONG_TYPE            = "pong"
	BOARD_CHANNEL_PREFIX = "board_"
)

var textChannel = make(chan *Message)
var databaseNotificationChannel = make(chan string)
var registerChannel = make(chan *WebsocketConnection)
var unregisterChannel = make(chan *WebsocketConnection)
var websocketConnections = WebsocketConnections{
	Connections: make(map[SessionId]*WebsocketConnection),
}
var websocketChannels = WebsocketChannels{
	Channels: make(map[ChannelName]WebsocketChannel),
}

func HandleUpgrade(c *fiber.Ctx) error {
	if !websocket.IsWebSocketUpgrade(c) {
		return fiber.ErrUpgradeRequired
	}

	return c.Next()
}

var HandleSocket = websocket.New(func(connection *websocket.Conn) {
	sessionId, ok := connection.Locals("sessionId").(SessionId)
	if !ok {
		connection.Close()
		return
	}

	user, ok := connection.Locals("user").(queries.User)
	if !ok {
		connection.Close()
		return
	}

	pongChannel := make(PongChannel, 1)
	closeChannel := make(CloseChannel, 1)

	websocketConnection := &WebsocketConnection{
		SessionId:    sessionId,
		User:         user,
		Connection:   connection,
		PongChannel:  &pongChannel,
		CloseChannel: &closeChannel,
	}

	registerChannel <- websocketConnection
	defer func() {
		websocketConnection.WaitGroup.Wait()
		websocketConnection.close()
	}()

	go websocketConnection.handlePingPong()

	for {
		websocketMessageType, message, err := websocketConnection.Connection.ReadMessage()
		if err != nil {
			break
		}

		switch websocketMessageType {
		case websocket.TextMessage:
			var unmarshaledMessage MessageContent
			if err := json.Unmarshal(message, &unmarshaledMessage); err != nil {
				continue
			}

			messageType, ok := unmarshaledMessage["type"].(string)
			if !ok {
				continue
			}

			switch messageType {
			case PING_TYPE:
				websocketConnection.writeMessage(PONG_TYPE, MessageContent{})
			case PONG_TYPE:
				select {
				case *websocketConnection.PongChannel <- struct{}{}:
				default:
				}
			default:
				channel, ok := unmarshaledMessage["channel"].(string)
				if !ok {
					continue
				}

				textChannel <- &Message{
					Channel:             channel,
					MessageType:         messageType,
					Message:             unmarshaledMessage,
					WebsocketConnection: websocketConnection,
				}
			}
		}
	}
}, websocket.Config{
	HandshakeTimeout: 10 * time.Second,
	ReadBufferSize:   2048,
	WriteBufferSize:  2048,
	Origins:          strings.Split(os.Getenv("ALLOWED_ORIGINS"), ","),
})

func Process() {
	conn, release, ctx, err := database.Acquire()
	if err != nil {
		return
	}
	defer release()

	conn.Exec(ctx, "LISTEN events")
	if err != nil {
		return
	}

	for {
		select {
		case databaseNotification := <-databaseNotificationChannel:
			// TODO
		case message := <-textChannel:
			switch message.MessageType {
			case JOIN_TYPE:
				if !message.WebsocketConnection.isAllowedToJoinChannel(message.Channel) {
					continue
				}

				websocketChannels.add(message.WebsocketConnection, message.Channel)

				websocketConnections.writeChannelMessage(message.Channel, message.MessageType, MessageContent{
					"userId": message.WebsocketConnection.User.ID,
				})
			case LEAVE_TYPE:
				if !message.WebsocketConnection.isInChannel(message.Channel) {
					continue
				}

				websocketChannels.remove(message.WebsocketConnection, message.Channel)

				websocketConnections.writeChannelMessage(message.Channel, message.MessageType, MessageContent{
					"userId": message.WebsocketConnection.User.ID,
				})
			}
		case websocketConnection := <-registerChannel:
			websocketConnections.writeGlobalMessage(REGISTER_TYPE, MessageContent{
				"userId": websocketConnection.User.ID,
			})

			websocketConnections.add(websocketConnection)
		case websocketConnection := <-unregisterChannel:
			for channel := range websocketChannels.Channels {
				if !websocketConnection.isInChannel(channel) {
					continue
				}

				websocketChannels.remove(websocketConnection, channel)

				websocketConnections.writeChannelMessage(channel, LEAVE_TYPE, MessageContent{
					"userId": websocketConnection.User.ID,
				})
			}

			websocketConnection.Connection.Close()

			websocketConnections.remove(websocketConnection)

			websocketConnections.writeGlobalMessage(UNREGISTER_TYPE, MessageContent{
				"userId": websocketConnection.User.ID,
			})
		}
	}
}
