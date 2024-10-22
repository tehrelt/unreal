package authservice

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/tehrelt/unreal/internal/dto"
	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/lib/imap"
	"github.com/tehrelt/unreal/internal/storage"
	"github.com/tehrelt/unreal/internal/storage/models"
)

func (s *AuthService) Login(ctx context.Context, in *dto.LoginDto) (*dto.LoginResult, error) {

	_, cleanup, err := imap.Dial(in.Email, in.Password, in.Imap.Host, in.Imap.Port)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	pass, err := s.encrypt(in.Password)
	if err != nil {
		return nil, fmt.Errorf("unable to encrypt password: %w", err)
	}

	out := &dto.LoginResult{
		FirstLogon: true,
	}

	info := &entity.SessionInfo{
		Email:    in.Email,
		Password: pass,
		Imap:     in.Imap,
		Smtp:     in.Smtp,
	}

	createModel := &models.CreateUser{
		UserBase: models.UserBase{
			Email: in.Email,
		},
	}
	if err := s.userSaver.Save(ctx, createModel); err != nil {
		if !errors.Is(err, storage.ErrUserAlreadyExists) {
			return nil, fmt.Errorf("unable to save user: %w", err)
		}
		out.FirstLogon = false
	}

	slog.Debug("creating session", slog.Any("session", info))

	id, err := s.sessions.Save(ctx, info, s.cfg.Jwt.Ttl)
	if err != nil {
		return nil, fmt.Errorf("unable to save session: %w", err)
	}

	slog.Debug("signing token with RSA")
	token, err := s.cfg.Jwt.RSA.Sign(&entity.Claims{Id: id}, s.cfg.Jwt.Ttl)
	if err != nil {
		return nil, fmt.Errorf("unable to sign token: %w", err)
	}

	out.Token = token

	return out, nil
}
