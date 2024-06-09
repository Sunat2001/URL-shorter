package jwt

import (
	"github.com/go-chi/jwtauth/v5"
	"os"
	"time"
)

var TokenAuth *jwtauth.JWTAuth

func Init() {
	TokenAuth = jwtauth.New(os.Getenv("JWT_ALGO"), []byte(os.Getenv("JWT_SECRET")), nil)
}

func GenerateToken(userId int64) (string, error) {
	if TokenAuth == nil {
		Init()
	}
	expirationTime := time.Now().Add(72 * time.Hour).Unix()

	claims := map[string]interface{}{
		"user_id": userId,
		"exp":     expirationTime,
	}

	_, tokenString, err := TokenAuth.Encode(claims)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
