package backend

import (
	"context"
	"fmt"

	pgxzap "github.com/jackc/pgx-zap"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"go.uber.org/zap"
)

type PostgresConfig struct {
	UserName string `mapstructure:"user_name" validate:"required"`
	Password string `mapstructure:"password" validate:"required"`
	Host     string `mapstructure:"host" validate:"required"`
	Database string `mapstructure:"database" validate:"required"`
	Port     string `mapstructure:"port" validate:"required"`
}

type PostgresBackend struct {
	Conn *pgxpool.Pool
}

func NewPostgresBackend(config PostgresConfig, logger *zap.SugaredLogger) (Backend, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", config.UserName, config.Password, config.Host, config.Port, config.Database)
	pgxconfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	pgxconfig.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   pgxzap.NewLogger(logger.Desugar()),
		LogLevel: tracelog.LogLevelInfo,
	}
	conn, err := pgxpool.NewWithConfig(context.Background(), pgxconfig)
	if err != nil {
		return nil, err
	}
	return &PostgresBackend{Conn: conn}, nil
}

func (b *PostgresBackend) GetCurrent(ctx context.Context, path string) (string, error) {
	result, err := b.GetManyCurrent(ctx, path)
	if err != nil {
		return "", err
	}
	if len(result) < 1 {
		return "", &NotFoundErr{Path: path}
	}
	return result[0], nil
}

func (b *PostgresBackend) Get(ctx context.Context, path string, version int) (string, error) {
	result, err := b.GetMany(ctx, path, version)
	if err != nil {
		return "", err
	}
	if len(result) < 1 {
		return "", &NotFoundErr{Path: path}
	}
	return result[0], nil
}

func (b *PostgresBackend) GetManyCurrent(ctx context.Context, path string) ([]string, error) {
	rows, err := b.Conn.Query(
		ctx,
		`SELECT config.value FROM config 
		INNER JOIN config_metadata ON config.path = config_metadata.path AND config.version = config_metadata.current_version
		WHERE config.path = $1`,
		path,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rows: %w", err)
	}
	defer rows.Close()
	result := []string{}
	for rows.Next() {
		var value string
		err = rows.Scan(&value)
		if err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, nil
}

func (b *PostgresBackend) GetMany(ctx context.Context, path string, version int) ([]string, error) {
	rows, err := b.Conn.Query(ctx, `SELECT "value" FROM "config" WHERE "path" = $1 AND "version"=$2`, path, version)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rows: %w", err)
	}
	defer rows.Close()
	result := []string{}
	for rows.Next() {
		var value string
		err = rows.Scan(&value)
		if err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, nil
}

func (b *PostgresBackend) Set(ctx context.Context, path, value string, options SetOptions) (*Metadata, error) {
	return b.SetMany(ctx, path, []string{value}, options)
}

func (b *PostgresBackend) SetMany(ctx context.Context, path string, values []string, options SetOptions) (*Metadata, error) {
	_, err := b.Conn.Exec(ctx, `INSERT INTO config_metadata (path) VALUES ($1) ON CONFLICT (path) DO NOTHING`, path)
	if err != nil {
		return nil, err
	}

	// the statements are wrapped inside a transaction to ensure the data insertion and metadata update is atomic
	// it also holds other connection from modifying the metadata entry
	tx, err := b.Conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var metadata Metadata
	row := tx.QueryRow(ctx, `UPDATE config_metadata SET latest_version = latest_version + 1, updated_at = NOW() WHERE path = $1 RETURNING (path, latest_version, current_version, created_at, updated_at)`, path)
	if err := row.Scan(&metadata); err != nil {
		return nil, err
	}

	if !options.KeepCurrent {
		row := tx.QueryRow(ctx, `UPDATE config_metadata SET current_version = latest_version WHERE path = $1 RETURNING (current_version)`, path)
		if err := row.Scan(&metadata.CurrentVersion); err != nil {
			return nil, err
		}
	}
	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"config"},
		[]string{"path", "version", "value"},
		pgx.CopyFromSlice(len(values), func(i int) ([]any, error) {
			return []any{path, metadata.LatestVersion, values[i]}, nil
		}),
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &metadata, nil
}

func (b *PostgresBackend) Delete(ctx context.Context, path string, version int) error {
	_, err := b.Conn.Exec(ctx, `DELETE FROM "config" WHERE "path" = $1 AND "version" = $2`, path, version)
	if err != nil {
		return fmt.Errorf("failed to delete rows: %w", err)
	}
	//TODO: change meta
	return nil
}

func (b *PostgresBackend) GetMetadata(ctx context.Context, path string) (*Metadata, error) {
	var metadata Metadata
	row := b.Conn.QueryRow(ctx, `SELECT path, latest_version, current_version, created_at, updated_at FROM config_metadata WHERE path = $1`, path)
	if err := row.Scan(&metadata.Path, &metadata.LatestVersion, &metadata.CurrentVersion, &metadata.CreatedAt, &metadata.CreatedAt); err != nil {
		return nil, err
	}
	return &metadata, nil
}

func (b *PostgresBackend) Close(context.Context) error {
	b.Conn.Close()
	return nil
}
