package websocket

import (
	"context"
	"sync"
	"time"

	"github.com/LeonardJouve/pass-secure/database"
	"github.com/LeonardJouve/pass-secure/database/queries"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type CloseChannel = chan struct{}

type Hub struct {
	closeChannel                CloseChannel
	databaseNotificationChannel DatabaseNotificationChannel
	sync.WaitGroup
}

func New() Hub {
	return Hub{
		closeChannel:                make(CloseChannel),
		databaseNotificationChannel: make(DatabaseNotificationChannel),
	}
}

func (hub *Hub) Close() {
	close(hub.closeChannel)
	hub.Wait()
}

func (hub *Hub) Process() {
	defer hub.Wait()

	hub.Add(1)
	go hub.listenDatabaseNotifications()

	for {
		select {
		case notification := <-hub.databaseNotificationChannel:
			// TODO
		case <-hub.closeChannel:
			return
		}
	}
}

func (hub *Hub) listenDatabaseNotifications() {
	defer hub.Done()

	conn, release, ctx, err := database.Acquire()
	if err != nil {
		return
	}
	defer release()

	if _, err := conn.Exec(ctx, "LISTEN websocket_events"); err != nil {
		return
	}

	ctx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup
	defer wg.Wait()
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			notification, err := conn.Conn().WaitForNotification(ctx)
			if err != nil {
				return
			}

			select {
			case hub.databaseNotificationChannel <- notification.Payload:
			default:
			}
		}
	}()

	for {
		select {
		case <-hub.closeChannel:
			cancel()
			return
		}
	}
}

func (hub *Hub) HandleUpgrade() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !websocket.IsWebSocketUpgrade(c) {
			return fiber.ErrUpgradeRequired
		}

		return c.Next()
	}
}

func (hub *Hub) HandleSocket() fiber.Handler {
	return websocket.New(func(connection *websocket.Conn) {
		user, ok := connection.Locals("user").(queries.User)
		if !ok {
			return
		}

		websocketConnection := WebsocketConnection{
			userId:       user.ID,
			connection:   connection,
			closeChannel: make(CloseChannel, 1),
			pongChannel:  make(PongChannel, 1),
		}

		websocketConnections.add(&websocketConnection)

		defer func() {
			websocketConnection.Wait()
			websocketConnection.close()
		}()

		websocketConnection.Add(1)
		go websocketConnection.handlePingPong()

		for {
			websocketMessageType, _, err := websocketConnection.connection.ReadMessage()
			if err != nil {
				break
			}

			switch websocketMessageType {
			case websocket.PongMessage:
				select {
				case websocketConnection.pongChannel <- struct{}{}:
					// TODO: received pong
				default:
				}
			case websocket.CloseMessage:
				close(websocketConnection.closeChannel)
				return
			}
		}
	}, websocket.Config{
		HandshakeTimeout: 10 * time.Second,
		// TODO Origins:          strings.Split(os.Getenv("ALLOWED_ORIGINS"), ","),
	})
}
