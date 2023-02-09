package app

import (
	"github.com/astaclinic/astafx"
	"go.uber.org/fx"

	"github.com/laminatedio/dendrite/internal/pkg/backend"
	"github.com/laminatedio/dendrite/internal/pkg/config"
	"github.com/laminatedio/dendrite/internal/pkg/dendrite"
)

func New() *fx.App {
	app := fx.New(
		astafx.Module,
		config.Module,
		dendrite.Module,
		backend.Module,
	)
	return app
}
