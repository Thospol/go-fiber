package jwt

import (
	"fmt"
	"time"

	"github.com/Thospol/go-fiber/internal/core/config"

	jwt "github.com/dgrijalva/jwt-go"
)

var (
	hmacSampleSecret []byte
)

// LoadKey load rsa key
func LoadKey() {
	hmacSampleSecret = []byte(config.CF.JWT.SecretKey)
}

// Signed signed payload with jwt RSA2048
func Signed(payload map[string]interface{}, exp time.Time) (string, error) {
	dataMap := jwt.MapClaims{
		"iat": time.Now().Unix(),
		"exp": exp.Unix(),
	}

	for key, value := range payload {
		dataMap[key] = value
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, dataMap)

	tokenString, err := token.SignedString(hmacSampleSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// Parsed parsed jwt token
func Parsed(tokenString string, onlyValid bool) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return hmacSampleSecret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && (!onlyValid || token.Valid) {
		result := make(map[string]interface{})
		for key, value := range claims {
			result[key] = value
		}
		return result, nil
	}

	return nil, config.RR.InvalidToken
}
