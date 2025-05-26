package websocket

import (
	"encoding/json"
	"time"

	"github.com/LeonardJouve/pass-secure/database"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type PongChannel = chan struct{}
type CloseChannel = chan struct{}
type DatabaseNotificationChannel = chan string
type MessageChannel = chan *Message

var websocketConnections = WebsocketConnections{
	Connections: make(map[SessionId]*WebsocketConnection),
}

func HandleUpgrade(c *fiber.Ctx) error {
	if !websocket.IsWebSocketUpgrade(c) {
		return fiber.ErrUpgradeRequired
	}

	return c.Next()
}

func makeWebsocketHandler(messageChannel MessageChannel) fiber.Handler {
	return websocket.New(func(connection *websocket.Conn) {
		// TODO: session id
		sessionId, ok := connection.Locals("sessionId").(SessionId)
		if !ok {
			connection.Close()
			return
		}

		websocketConnection := &WebsocketConnection{
			SessionId:    sessionId,
			Connection:   connection,
			CloseChannel: make(CloseChannel, 1),
		}

		websocketConnections.add(websocketConnection)

		defer func() {
			websocketConnection.WaitGroup.Wait()
			websocketConnection.close()
		}()

		pongChannel := make(PongChannel, 1)
		go websocketConnection.handlePingPong(pongChannel)

		for {
			websocketMessageType, message, err := websocketConnection.Connection.ReadMessage()
			if err != nil {
				break
			}

			switch websocketMessageType {
			case websocket.PingMessage:
				websocketConnection.Connection.WriteMessage(websocket.PongMessage, []byte{})
			case websocket.PongMessage:
				select {
				case pongChannel <- struct{}{}:
				default:
				}
			case websocket.CloseMessage:
				// TODO: now close connection
			case websocket.TextMessage:
				var unmarshaledMessage MessageContent
				if err := json.Unmarshal(message, &unmarshaledMessage); err != nil {
					continue
				}

				messageType, ok := unmarshaledMessage["type"].(string)
				if !ok {
					continue
				}

				select {
				case messageChannel <- &Message{
					MessageType:         messageType,
					Message:             unmarshaledMessage,
					WebsocketConnection: websocketConnection,
				}:
				default:
				}
			}
		}
	}, websocket.Config{
		HandshakeTimeout: 10 * time.Second,
		// TODO Origins:          strings.Split(os.Getenv("ALLOWED_ORIGINS"), ","),
	})
}

func Process() {
	// TODO: handle stop
	messageChannel := make(MessageChannel)
	databaseNotificationChannel := make(DatabaseNotificationChannel)

	go listenDatabaseNotifications(databaseNotificationChannel)

	for {
		select {
		case notification := <-databaseNotificationChannel:
			// TODO
		case message := <-messageChannel:
			// TODO
		}
	}
}

func listenDatabaseNotifications(databaseNotificationChannel DatabaseNotificationChannel) {
	conn, release, ctx, err := database.Acquire()
	if err != nil {
		return
	}
	defer release()

	// TODO: handle stop context.SetReadDeadline(ctx, )

	if _, err := conn.Exec(ctx, "LISTEN websocket_events"); err != nil {
		return
	}

	for {
		notification, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			return
		}

		select {
		case databaseNotificationChannel <- notification.Payload:
		default:
		}
	}
}
