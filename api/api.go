package api

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/LeonardJouve/pass-secure/auth"
	"github.com/LeonardJouve/pass-secure/status"
	"github.com/LeonardJouve/pass-secure/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/gofiber/storage/redis/v3"
)

func HealthCheck(c *fiber.Ctx) error {
	return status.Ok(c, nil)
}

func Start(port uint16) (func() error, error) {
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins:     os.Getenv("ALLOWED_ORIGINS"),
		AllowHeaders:     "Origin, Content-Type, Accept, X-CSRF-Token, Authorization",
		AllowMethods:     "GET, POST, PUT, PATCH, DELETE",
		AllowCredentials: true,
	}))

	csrfTokenExpirationString := os.Getenv("CSRF_TOKEN_LIFETIME_IN_MINUTE")
	csrfTokenExpiration, err := strconv.ParseInt(csrfTokenExpirationString, 10, 64)
	if err != nil {
		return nil, err
	}

	redisPortString := os.Getenv("REDIS_PORT")
	redisPort, err := strconv.ParseInt(redisPortString, 10, 64)
	if err != nil {
		return nil, err
	}

	app.Use(csrf.New(csrf.Config{
		ContextKey:     auth.CSRF_TOKEN,
		CookieName:     auth.CSRF_TOKEN,
		CookieDomain:   os.Getenv("HOST"),
		CookiePath:     "/",
		CookieSecure:   true,
		CookieSameSite: "Lax",
		CookieHTTPOnly: true,
		Extractor:      auth.CsrfTokenExtractor,
		Expiration:     time.Duration(csrfTokenExpiration) * time.Minute,
		KeyGenerator:   utils.UUIDv4,
		Storage: redis.New(redis.Config{
			Host:     os.Getenv("REDIS_HOST"),
			Port:     int(redisPort),
			Password: os.Getenv("REDIS_PASSWORD"),
		}),
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return status.Unauthorized(c, errors.New("invalid csrf token"))
		},
	}))

	app.Get("/healthcheck", HealthCheck)

	app.Post("/login", Login)
	app.Post("/register", Register)

	apiGroup := app.Group("", Protect)

	websocketTimeoutString := os.Getenv("WEBSOCKET_TIMEOUT_IN_SECOND")
	websocketTimeout, err := strconv.ParseInt(websocketTimeoutString, 10, 64)
	if err != nil {
		return nil, err
	}

	hub := websocket.New(time.Duration(websocketTimeout) * time.Second)
	go hub.Process()
	apiGroup.Get("/ws", hub.HandleUpgrade(), hub.HandleSocket())

	folderGroup := apiGroup.Group("/folders")
	folderGroup.Get("/", GetFolders)
	folderGroup.Get("/:folder_id", GetFolder)
	folderGroup.Post("/", CreateFolder)
	folderGroup.Put("/:folder_id", UpdateFolder)
	folderGroup.Delete("/:folder_id", RemoveFolder)

	entriesGroup := apiGroup.Group("/entries")
	entriesGroup.Get("/", GetEntries)
	entriesGroup.Get("/:entry_id", GetEntry)
	entriesGroup.Post("/", CreateEntry)
	entriesGroup.Put("/:entry_id", UpdateEntry)
	entriesGroup.Delete("/:entry_id", RemoveEntry)

	usersGroup := apiGroup.Group("/users")
	usersGroup.Get("/", GetUsers)
	usersGroup.Get("/me", GetMe)
	usersGroup.Delete("/me", RemoveMe)
	usersGroup.Put("/me", UpdateMe)
	usersGroup.Get("/:user_id", GetUser)

	app.Listen(fmt.Sprintf(":%d", port))

	return func() error {
		hub.Close()

		return app.Shutdown()
	}, nil
}
