package backend

import (
	"context"

	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewBackend),
	fx.Invoke(func(lc fx.Lifecycle, b Backend) error {
		lc.Append(fx.Hook{
			OnStop: func(ctx context.Context) error {
				return b.Close(ctx)
			},
		})
		return nil
	}),
)
