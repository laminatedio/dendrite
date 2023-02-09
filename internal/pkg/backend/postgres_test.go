package backend_test

import (
	"context"
	_ "embed"
	"errors"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/laminatedio/dendrite/internal/pkg/backend"
)

//go:embed schema/config.sql
var postgresSchema string

var _ = Describe("Postgres", func() {
	var conn *pgxpool.Pool
	var pgBackend *backend.PostgresBackend

	BeforeEach(func(ctx context.Context) {
		var err error
		conn, err = pgxpool.New(ctx, os.Getenv("POSTGRES_DSN"))
		Expect(err).ShouldNot(HaveOccurred())
		pgBackend = &backend.PostgresBackend{
			Conn: conn,
		}
		_, err = conn.Exec(ctx, "DROP SCHEMA public CASCADE")
		Expect(err).ShouldNot(HaveOccurred())
		_, err = conn.Exec(ctx, "CREATE SCHEMA IF NOT EXISTS public")
		Expect(err).ShouldNot(HaveOccurred())
		_, err = conn.Exec(ctx, postgresSchema)
		Expect(err).ShouldNot(HaveOccurred())
	})

	Describe("Get", func() {
		When("there is no metadata of the path exists", func() {
			It("should return not found error", func(ctx context.Context) {
				_, err := pgBackend.Get(ctx, "/some/nonexist/path", 3)
				var notFoundErr *backend.NotFoundErr
				Expect(errors.As(err, &notFoundErr)).To(BeTrue())
				Expect(notFoundErr.Path).To(Equal("/some/nonexist/path"))
			})
		})
		When("there is a single value stored in the path and version", func() {
			path := "/some/test/path"
			value := "some_test_value"
			It("should return the value", func(ctx context.Context) {
				Expect(conn.Exec(ctx, "INSERT INTO config_metadata (path, latest_version, current_version) VALUES ($1, 2, 2)", path)).Error().NotTo(HaveOccurred())
				Expect(conn.Exec(ctx, "INSERT INTO config (path, version, value) VALUES ($1, 2, $2)", path, value)).Error().NotTo(HaveOccurred())
				actualValue, err := pgBackend.Get(ctx, path, 2)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualValue).To(Equal(value))
			})
		})
	})

	AfterEach(func(ctx context.Context) {
		Expect(pgBackend.Close(ctx)).To(Succeed())
	})
})
