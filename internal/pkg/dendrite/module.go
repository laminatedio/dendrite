package dendrite

import (
	"github.com/astaclinic/astafx/routerfx"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(routerfx.AsControllerRoute(NewDendriteController)),
	fx.Provide(NewDendriteService),
)
