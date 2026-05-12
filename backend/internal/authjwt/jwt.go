package authjwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"findus/backend/internal/domain"
)

const Issuer = "findus"

type Claims struct {
	UserID string `json:"uid"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func Sign(secret []byte, userID string, role domain.Role, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Role:   string(role),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(secret)
}

func Parse(secret []byte, token string) (*Claims, error) {
	c, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := c.Claims.(*Claims)
	if !ok || !c.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
