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
type DatabaseNotificationChannel = chan string

type Hub struct {
	timeout                     time.Duration
	connections                 WebsocketConnections
	closeChannel                CloseChannel
	databaseNotificationChannel DatabaseNotificationChannel
	sync.WaitGroup
}

func New(timeout time.Duration) Hub {
	return Hub{
		timeout: timeout,
		connections: WebsocketConnections{
			connections: make(map[int64]UserConnections),
		},
		closeChannel:                make(CloseChannel),
		databaseNotificationChannel: make(DatabaseNotificationChannel),
	}
}

func (h *Hub) Close() {
	// TODO send or close ?
	close(h.closeChannel)
	h.Wait()
}

func (h *Hub) Process() {
	defer h.Wait()

	h.Add(1)
	go h.listenDatabaseNotifications()

	for {
		select {
		case notification := <-h.databaseNotificationChannel:
			// TODO
		case <-h.closeChannel:
			return
		}
	}
}

func (h *Hub) listenDatabaseNotifications() {
	defer h.Done()

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
			case h.databaseNotificationChannel <- notification.Payload:
			default:
			}
		}
	}()

	for {
		select {
		case <-h.closeChannel:
			cancel()
			return
		}
	}
}

func (h *Hub) HandleUpgrade() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !websocket.IsWebSocketUpgrade(c) {
			return fiber.ErrUpgradeRequired
		}

		return c.Next()
	}
}

func (h *Hub) HandleSocket() fiber.Handler {
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

		h.connections.add(&websocketConnection)

		defer func() {
			websocketConnection.Wait()
			websocketConnection.close()
		}()

		websocketConnection.Add(1)
		go websocketConnection.handlePingPong(h.timeout)

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
