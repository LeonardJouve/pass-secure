package auth

import (
	"fmt"
	"strconv"
	"time"

	"github.com/LeonardJouve/pass-secure/status"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/golang-jwt/jwt/v5"
)

const (
	ACCESS_TOKEN = "access_token"
)

func CreateToken(c *fiber.Ctx, userId uint) (string, bool) {
	privateKey, ok := getPrivateKey(c, ACCESS_TOKEN)
	if !ok {
		status.InternalServerError(c, nil)
		return "", false
	}

	jwt.TimePrecision = time.Microsecond
	claims := &jwt.RegisteredClaims{
		ID:       utils.UUIDv4(),
		Subject:  strconv.FormatUint(uint64(userId), 10),
		IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(privateKey)
	if err != nil {
		status.InternalServerError(c, nil)
		return "", false
	}

	c.Cookie(&fiber.Cookie{
		Name:     ACCESS_TOKEN,
		Value:    token,
		Path:     "/",
		Secure:   true,
		HTTPOnly: true,
	})

	return token, true
}

func ValidateToken(c *fiber.Ctx, name string, token string) bool {
	publicKey, ok := getPublicKey(c, name)
	if !ok {
		status.Unauthorized(c, nil)
		return false
	}

	var claims = jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})
	if err != nil {
		status.Unauthorized(c, nil)
		return false
	}

	return true
}
