package httpvalidator

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type EchoValidator struct {
	validator *validator.Validate
}

func New(validator *validator.Validate) *EchoValidator {
	return &EchoValidator{validator}
}

func (ev *EchoValidator) Validate(i any) error {
	if err := ev.validator.Struct(i); err != nil {
		return echo.NewHTTPError(echo.ErrBadRequest.Code, err.Error())
	}
	return nil
}
