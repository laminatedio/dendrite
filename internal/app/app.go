package app

import (
	"github.com/astaclinic/astafx"
	"go.uber.org/fx"

	"dendrite/internal/pkg/backend"
	"dendrite/internal/pkg/config"
	"dendrite/internal/pkg/dendrite"
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
