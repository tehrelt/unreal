package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/tehrelt/unreal/internal/entity"
)

type claims struct {
	entity.Claims
	jwt.RegisteredClaims
}
