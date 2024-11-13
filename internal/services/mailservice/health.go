package mailservice

import (
	"context"
	"fmt"

	"github.com/tehrelt/unreal/internal/entity"
)

func (s *Service) Health(ctx context.Context) (*entity.HealthInfo, error) {

	fn := "mailservice.Health"
	info := &entity.HealthInfo{
		Version: s.cfg.App.Version,
	}

	s.m.Do(ctx, func(ctx context.Context) error {
		enabled, err := s.r.Health(ctx)
		if err != nil {
			return fmt.Errorf("%s: %w", fn, err)
		}

		info.TlsEnabled = enabled

		return nil
	})

	return info, nil
}
