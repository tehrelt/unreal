package handlers

import (
	"mime/multipart"

	"github.com/labstack/echo/v4"
	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/services/authservice"
)

type UpdateProfileRequest struct {
	Name *string `json:"name,omitempty" form:"name"`
}

func UpdateProfile(s *authservice.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req UpdateProfileRequest

		ctx, u, err := extractUser(c)
		if err != nil {
			return err
		}

		form, err := c.MultipartForm()
		if err != nil {
			return err
		}

		if len(form.Value["name"]) > 0 {
			req.Name = &form.Value["name"][0]
		}

		files := form.File["picture"]
		var file *multipart.FileHeader
		if len(files) > 0 {
			file = files[0]
		}

		if err := s.UpdateUser(ctx, &entity.UpdateUser{
			Email:   u.Email,
			Name:    req.Name,
			Picture: file,
		}); err != nil {
			return err
		}

		return c.NoContent(200)
	}
}
