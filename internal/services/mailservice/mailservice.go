package mailservice

import (
	"github.com/tehrelt/unreal/internal/config"
)

type MailService struct {
	cfg *config.Config
}

func New(cfg *config.Config) *MailService {
	return &MailService{cfg: cfg}
}
