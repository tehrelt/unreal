package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/services/authservice"
)

type UpdateProfileRequest struct {
	Name *string `json:"name,omitempty"`
}

func UpdateProfile(s *authservice.AuthService) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req UpdateProfileRequest

		ctx, u, err := extractUser(c)
		if err != nil {
			return err
		}

		if err := c.Bind(&req); err != nil {
			return err
		}

		if err := s.UpdateUser(ctx, &entity.UpdateUser{
			Email:   u.Email,
			Name:    req.Name,
			Picture: nil,
		}); err != nil {
			return err
		}

		return c.NoContent(200)
	}
}
