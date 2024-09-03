//go:build wireinject
// +build wireinject

package app

import (
	"github.com/google/wire"
	"github.com/tehrelt/unreal/internal/config"
	"github.com/tehrelt/unreal/internal/services/authservice"
)

func New() (*App, error) {
	panic(wire.Build(
		newApp,
		config.New,
		authservice.New,
	))
}
