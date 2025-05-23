package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/LeonardJouve/pass-secure/database"
	"github.com/LeonardJouve/pass-secure/status"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/golang-jwt/jwt/v5"
)

const (
	ACCESS_TOKEN = "access_token"
)

func CreateToken(c *fiber.Ctx, userId int64) (string, bool) {
	privateKey, ok := getPrivateKey(c, ACCESS_TOKEN)
	if !ok {
		status.InternalServerError(c, nil)
		return "", false
	}

	jwt.TimePrecision = time.Microsecond
	claims := &jwt.RegisteredClaims{
		ID:       utils.UUIDv4(),
		Subject:  strconv.FormatInt(userId, 10),
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

func ValidateToken(c *fiber.Ctx, token string) (jwt.RegisteredClaims, bool) {
	publicKey, ok := getPublicKey(c, ACCESS_TOKEN)
	if !ok {
		status.Unauthorized(c, nil)
		return jwt.RegisteredClaims{}, false
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
		return jwt.RegisteredClaims{}, false
	}

	return claims, true
}

func IsExpired(c *fiber.Ctx, claims jwt.RegisteredClaims) bool {
	qtx, ctx, commit, ok := database.BeginTransaction(c)
	if !ok {
		return true
	}
	defer commit()

	userId, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		status.InternalServerError(c, nil)
		return true
	}

	// TODO
	_, err = qtx.GetUser(ctx, userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			status.Unauthorized(c, nil)
			return true
		} else {
			status.InternalServerError(c, nil)
			return true
		}
	}

	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now().UTC()) {
		status.Unauthorized(c, nil)
		return true
	}

	return false
}
