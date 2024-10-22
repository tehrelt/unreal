package smtp

import (
	"log/slog"

	"github.com/tehrelt/unreal/internal/config"
	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"gopkg.in/gomail.v2"
)

type Repository struct {
	cfg    *config.Config
	ctxman *smtpCtxManager
	l      *slog.Logger
}

func NewRepository(cfg *config.Config) *Repository {
	return &Repository{
		cfg:    cfg,
		ctxman: defaultManager,
		l:      slog.With(sl.Module("mail.SmtpRepository")),
	}
}

func (r *Repository) dial(session *entity.SessionInfo) (*gomail.Dialer, error) {
	fn := "context.dial"
	log := r.l.With(sl.Method(fn))

	log.Debug("dialing smtp")
	dialer := gomail.NewDialer(
		session.Smtp.Host,
		session.Smtp.Port,
		session.Email,
		session.Password,
	)

	return dialer, nil
}
