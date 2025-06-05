package websocket

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/LeonardJouve/pass-secure/database"
	"github.com/LeonardJouve/pass-secure/database/queries"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
)

type CloseChannel = chan struct{}
type DatabaseNotificationChannel = chan string

type Hub struct {
	timeout                     time.Duration
	connections                 WebsocketConnections
	closeChannel                CloseChannel
	databaseNotificationChannel DatabaseNotificationChannel
	sync.WaitGroup
	sync.Once
}

type Notification struct {
	Message   json.RawMessage `json:"message"`
	Broadcast bool            `json:"broadcast"`
	UserIds   []int64         `json:"user_ids"`
}

const (
	MAX_READ_SIZE       = 512
	WRITE_TIMEOUT       = 3 * time.Second
	WRITE_WORKER_AMOUNT = 5
)

func New(timeout time.Duration) Hub {
	return Hub{
		timeout: timeout,
		connections: WebsocketConnections{
			connections:  make(map[int64]UserConnections),
			writeChannel: make(WriteChannel, WRITE_WORKER_AMOUNT),
			closeChannel: make(CloseChannel, 1),
		},
		closeChannel:                make(CloseChannel),
		databaseNotificationChannel: make(DatabaseNotificationChannel),
	}
}

func (h *Hub) Close() {
	h.Do(func() {
		close(h.closeChannel)
		h.Wait()
		close(h.databaseNotificationChannel)
		h.connections.close()
	})
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
	defer cancel()

	h.Add(1)
	go func() {
		defer h.Done()

		for {
			notification, err := conn.Conn().WaitForNotification(ctx)
			if err != nil {
				return
			}

			select {
			case h.databaseNotificationChannel <- notification.Payload:
			case <-ctx.Done():
				return
			}
		}
	}()

	<-h.closeChannel
}

func (h *Hub) Process() {
	defer h.Wait()

	h.Add(1)
	go h.listenDatabaseNotifications()

	h.connections.Add(1)
	go h.connections.handleWriteWorkers()

	for {
		select {
		case databaseNotification := <-h.databaseNotificationChannel:
			var notification Notification
			if err := json.Unmarshal([]byte(databaseNotification), &notification); err != nil {
				continue
			}

			h.connections.sendNotification(notification)
		case <-h.closeChannel:
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
			id:                     utils.UUIDv4(),
			userId:                 user.ID,
			connection:             connection,
			closeChannel:           make(CloseChannel, 1),
			closeGracefullyChannel: make(CloseChannel, 1),
			writeTimeout:           WRITE_TIMEOUT,
		}
		defer websocketConnection.close()

		h.connections.add(&websocketConnection)
		defer h.connections.remove(&websocketConnection)

		websocketConnection.connection.SetReadLimit(MAX_READ_SIZE)

		websocketConnection.Add(1)
		go websocketConnection.handlePingPong(h.timeout)

		websocketConnection.Add(1)
		go websocketConnection.readMessages()

		for {
			select {
			case <-h.closeChannel:
				return
			case <-websocketConnection.closeChannel:
				return
			}
		}
	}, websocket.Config{
		HandshakeTimeout: 10 * time.Second,
		Origins:          strings.Split(os.Getenv("ALLOWED_ORIGINS"), ","),
	})
}
