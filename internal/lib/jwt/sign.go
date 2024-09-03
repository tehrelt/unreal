package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tehrelt/unreal/internal/entity"
)

func (j *JWT) Sign(claim *entity.Claims, ttl time.Duration) (string, error) {
	payload := claims{
		Claims: *claim,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(j.Private)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodPS256.SigningMethodRSA, payload)

	signed, err := token.SignedString(key)
	if err != nil {
		return "", err
	}

	return signed, nil
}
