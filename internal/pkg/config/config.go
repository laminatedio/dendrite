package config

import (
	"dendrite/internal/pkg/backend"
	"fmt"
	"os"

	"github.com/astaclinic/astafx/httpfx"
	"github.com/astaclinic/astafx/loggerfx"
	"github.com/astaclinic/astafx/sentryfx"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"gopkg.in/yaml.v3"
)

type Config struct {
	fx.Out  `yaml:"-"`
	Http    *httpfx.HttpConfig     `validate:"required"`
	Logs    *loggerfx.LoggerConfig `validate:"required"`
	Sentry  *sentryfx.SentryConfig `validate:"required"`
	Backend *backend.Config        `validate:"required"`
}

func NewConfig(validate *validator.Validate) (Config, error) {
	var config Config

	if err := viper.Unmarshal(&config); err != nil {
		return Config{}, fmt.Errorf("fail to unmarshal config: %w", err)
	}

	if err := validate.Struct(&config); err != nil {
		return Config{}, fmt.Errorf("config is invalid: %w", err)
	}

	return config, nil
}

func PrintConfig() error {
	config, err := NewConfig(validator.New())
	if err != nil {
		return err
	}
	configYaml, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(configYaml)
	if err != nil {
		return err
	}
	return nil
}

var Module = fx.Options(
	fx.Provide(validator.New),
	fx.Provide(NewConfig),
)
