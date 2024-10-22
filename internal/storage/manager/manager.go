package manager

import (
	"context"

	"github.com/tehrelt/unreal/internal/storage"
	"github.com/tehrelt/unreal/internal/storage/imap"
	"github.com/tehrelt/unreal/internal/storage/smtp"
)

type Manager struct {
	smtp storage.Manager
	imap storage.Manager
}

func NewManager() storage.Manager {
	return &Manager{
		imap: imap.NewManager(),
		smtp: smtp.NewManager(),
	}
}

func (m *Manager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return m.imap.Do(ctx, func(ctx context.Context) error {
		return m.smtp.Do(ctx, func(ctx context.Context) error {
			return fn(ctx)
		})
	})
}
