package imap

import (
	"context"
	"fmt"
)

func (r *Repository) Health(ctx context.Context) (bool, error) {

	fn := "Repository.Health"

	c, err := r.ctxman.get(ctx)
	if err != nil {
		return false, fmt.Errorf("%s: %w", fn, err)
	}

	return c.IsTLS(), nil
}
