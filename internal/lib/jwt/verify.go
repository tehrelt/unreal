package jwt

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tehrelt/unreal/internal/entity"
)

func (j *JWT) Verify(tokenString string) (*entity.Claims, error) {

	key, err := jwt.ParseRSAPublicKeyFromPEM(j.Public)
	if err != nil {
		return nil, fmt.Errorf("unable parse public key: %w", err)
	}

	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return key, nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable parse jwt token: %w", err)
	}

	claims, ok := token.Claims.(*claims)
	if !ok {
		return nil, fmt.Errorf("unable parse claims: %w", err)
	}
	if !token.Valid {
		return nil, ErrTokenExpired
	}

	return &claims.Claims, nil
}
