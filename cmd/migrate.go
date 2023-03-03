/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"errors"
	"fmt"

	astafxConfig "github.com/astaclinic/astafx/config"
	"github.com/astaclinic/astafx/logger"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laminatedio/dendrite/internal/pkg/backend"
	"github.com/laminatedio/dendrite/internal/pkg/backend/schema"
	"github.com/laminatedio/dendrite/internal/pkg/config"
	"github.com/spf13/cobra"
)

func newPostgresConn(config backend.PostgresConfig, root bool) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s", config.UserName, config.Password, config.Host, config.Port)
	if !root {
		dsn += fmt.Sprintf("/%s", config.Database)
	}
	pgxconfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	return pgxpool.NewWithConfig(context.Background(), pgxconfig)
}

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migration to postgres db",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Starting for postgres DB migration ...")
		ctx := context.Background()

		logger.Info("Getting config ...")
		astafxConfig.InitConfig(cfgFile)
		config, err := config.GetConfig()
		if err != nil {
			logger.Fatalf("fail to get config: %v", err.Error())
			return
		}

		logger.Info("Connecting to postgres ...")
		db, err := newPostgresConn(config.Backend.Postgres, true)
		if err != nil {
			logger.Fatalf("fail to get db conn: %v", err.Error())
			return
		}
		defer db.Close()

		logger.Info("Create database if not exist ...")
		database := config.Backend.Postgres.Database
		_, err = db.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", database))
		if err != nil {
			var e *pgconn.PgError
			// if the error is database existed
			if errors.As(err, &e) && e.Code == "42P04" {
				logger.Info("The database is already existed")
			} else {
				logger.Fatalf("fail to init db: %v", err.Error())
				return
			}
		}

		logger.Info("Connecting to target postgres db ...")
		dendriteDb, err := newPostgresConn(config.Backend.Postgres, false)
		if err != nil {
			logger.Fatalf("fail to get db conn: %v", err.Error())
			return
		}
		defer dendriteDb.Close()

		logger.Info("Create table if not exist ...")
		schema := schema.GetSchema()
		_, err = dendriteDb.Exec(ctx, schema)
		if err != nil {
			logger.Fatalf("fail to init table: %v", err.Error())
			return
		}

		logger.Info("Migrate successfully")
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// migrateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// migrateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
