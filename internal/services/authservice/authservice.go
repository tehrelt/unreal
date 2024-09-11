package authservice

import (
	"github.com/tehrelt/unreal/internal/config"
)

type AuthService struct {
	cfg *config.Config
}

func New(cfg *config.Config) *AuthService {
	return &AuthService{cfg: cfg}
}
