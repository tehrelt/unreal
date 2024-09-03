package authservice

import (
	"context"
	"crypto/aes"
	"encoding/hex"
	"fmt"
	"log/slog"

	"github.com/tehrelt/unreal/internal/config"
	"github.com/tehrelt/unreal/internal/dto"
	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/lib/imap"
)

type AuthService struct {
	cfg *config.Config
}

func New(cfg *config.Config) *AuthService {
	return &AuthService{cfg: cfg}
}

func (s *AuthService) Login(ctx context.Context, d *dto.LoginDto) (string, error) {

	_, cleanup, err := imap.Dial(d.Email, d.Password, d.Host, d.Port)
	if err != nil {
		return "", err
	}
	defer cleanup()

	slog.Debug("dial successfull")

	ciph, err := aes.NewCipher([]byte(s.cfg.AES.Secret))
	if err != nil {
		return "", err
	}

	slog.Debug("ciphering aes")
	buf := make([]byte, len(d.Password))
	ciph.Encrypt(buf, []byte(d.Password))
	encpass := hex.EncodeToString(buf)

	claims := &entity.Claims{
		Email:    d.Email,
		Host:     d.Host,
		Password: encpass,
		Port:     d.Port,
	}

	slog.Debug("signing token with RSA")
	token, err := s.cfg.Jwt.RSA.Sign(claims, s.cfg.Jwt.Ttl)
	if err != nil {
		return "", fmt.Errorf("unable to sign token: %w", err)
	}

	slog.Info("signed token", slog.String("token", token))

	return token, nil
}
