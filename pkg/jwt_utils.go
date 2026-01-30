package pkg

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var jwtSecret []byte

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string    `json:"role"` // role nya harus "USER, MODERATOR"
	jwt.RegisteredClaims
}

func JwtInit(secret string) {
	jwtSecret = []byte(secret)
	if len(jwtSecret) == 0 {
		panic("JWT_SECRET not set")
	}
}

func GenerateToken(userID uuid.UUID, role string, limit int) (string, error) {
	expirationTime := time.Now().Add(time.Duration(limit) * time.Minute)

	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "agmer-yapping",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
