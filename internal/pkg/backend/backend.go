package backend

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type NotFoundErr struct {
	Path string
}

func (err *NotFoundErr) Error() string {
	return fmt.Sprintf("path %s not found", err.Path)
}

type Record struct {
	Value string
	Metadata
}

type Metadata struct {
	Path           string
	LatestVersion  int
	CurrentVersion int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type SetOptions struct {
	KeepCurrent bool
}

type SetMetaDataOptions struct {
	CurrentVersion int
	LatestVersion  int
}

type Backend interface {
	GetCurrent(ctx context.Context, path string) (string, error)
	Get(ctx context.Context, path string, version int) (string, error)
	GetManyCurrent(ctx context.Context, path string) ([]string, error)
	GetMany(ctx context.Context, path string, version int) ([]string, error)
	Set(ctx context.Context, path string, value string, options SetOptions) (*Metadata, error)
	SetMany(ctx context.Context, path string, values []string, options SetOptions) (*Metadata, error)
	Close(ctx context.Context) error
}

type Config struct {
	Type     string         `mapstructure:"type" validate:"required"`
	Postgres PostgresConfig `mapstructure:"postgres"`
}

func NewBackend(config *Config, logger *zap.SugaredLogger) (Backend, error) {
	switch config.Type {
	case "postgres":
		return NewPostgresBackend(config.Postgres, logger)
	}
	return nil, fmt.Errorf("backend %s not implemented", config.Type)
}
