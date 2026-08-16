package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func Generate(userID, role, secret string, hours int) (string, error) {
	now := time.Now()
	return jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{UserID: userID, Role: role, RegisteredClaims: jwt.RegisteredClaims{Subject: userID, IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(hours) * time.Hour))}}).SignedString([]byte(secret))
}
func Parse(token, secret string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !parsed.Valid {
		if err == nil {
			err = jwt.ErrTokenInvalidClaims
		}
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || claims.UserID == "" || (claims.Role != "admin" && claims.Role != "user") {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}
