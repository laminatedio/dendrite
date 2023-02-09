package dendrite

import (
	"context"
	"log"
	"reflect"
	"testing"

	"github.com/laminatedio/dendrite/internal/pkg/backend"
	"github.com/laminatedio/dendrite/internal/pkg/dendrite/dto"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tmc/graphql"
)

type Config struct {
	Path  string
	Value string
}

type MetaData struct {
	Path           string
	LatestVersion  int
	CurrentVersion int
}

const postgresDsn = "postgres://postgres:mysecretpassword@localhost:5432/config"

var testConfig = []Config{
	{
		Path:  "/A/B/C",
		Value: "1",
	},
	{
		Path:  "/A/B/D",
		Value: "2",
	},
	{
		Path:  "/A/B",
		Value: "C",
	},
	{
		Path:  "/A/B",
		Value: "D",
	},
}

func ImportTestData(ctx context.Context, configs []Config, worker backend.Backend) error {
	for _, config := range configs {
		_, err := worker.Set(ctx, config.Path, config.Value, backend.SetOptions{KeepCurrent: false})
		if err != nil {
			return err
		}
	}
	return nil
}

func CleanUp(ctx context.Context) error {
	_, err := postgres.Conn.Exec(ctx, "DELETE FROM config")
	if err != nil {
		return err
	}
	_, err = postgres.Conn.Exec(ctx, `DELETE FROM config_metadata`)
	if err != nil {
		return err
	}
	return nil
}

func GetPostgresBackend() backend.PostgresBackend {
	conn, err := pgxpool.New(context.Background(), postgresDsn)
	if err != nil {
		log.Fatalf("fail to init db client: %v", err)
	}
	return backend.PostgresBackend{
		Conn: conn,
	}
}

var postgres backend.PostgresBackend = GetPostgresBackend()
var s = DendriteService{
	backend: &postgres,
}

// integration test (need db)
func TestDendriteService_Query(t *testing.T) {
	type args struct {
		ctx   context.Context
		query string
	}
	tests := []struct {
		name    string
		configs []Config
		args    args
		want    map[string]any
		wantErr bool
	}{
		{
			name:    "get data by query",
			configs: testConfig,
			args: args{
				ctx: context.Background(),
				query: `{
					A {
						B {
							C
							D
						}
					}
					A {
						B
					}
				}`,
			},
			want: map[string]any{
				"A": map[string]any{
					"B": map[string]any{
						"C": "1",
						"D": "2",
						"/": []string{"C", "D"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "get version 2",
			configs: []Config{
				{
					Path:  "/A",
					Value: "1",
				},
				{
					Path:  "/A",
					Value: "2",
				},
			},
			args: args{
				ctx: context.Background(),
				query: `{
					A (version:1)
				}`,
			},
			want: map[string]any{
				"A": "1",
			},
			wantErr: false,
		},
		{
			name:    "some key not found",
			configs: testConfig,
			args: args{
				ctx: context.Background(),
				query: `{
					A {
						B {
							C
							D
						}
					}
					D
				}`,
			},
			want: map[string]any{
				"A": map[string]any{
					"B": map[string]any{
						"C": "1",
						"D": "2",
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "invalid query",
			configs: testConfig,
			args: args{
				ctx:   context.Background(),
				query: `haha`,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fail := false
			err := ImportTestData(context.Background(), tt.configs, &postgres)
			if err != nil {
				t.Errorf("failed to import data: %v", err)
				return
			}
			got, err := s.Query(tt.args.ctx, tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("DendriteService.Query() error = %v, wantErr %v", err, tt.wantErr)
				fail = true
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DendriteService.Query() = %#v, want %#v", got, tt.want)
				fail = true
			}
			err = CleanUp(context.Background())
			if err != nil {
				t.Errorf("failed to clean up: %v", err)
				return
			}
			if fail {
				return
			}
		})
	}
}

func TestDendriteService_GetFieldVersion(t *testing.T) {
	type fields struct {
		postgres *pgx.Conn
	}
	type args struct {
		args []graphql.Argument
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		testingData map[string]string
		want        int
		wantErr     bool
	}{
		// TODO: Add test cases.
		{
			name: "should get version in args",
			fields: fields{
				postgres: nil,
			},
			args: args{
				args: []graphql.Argument{
					{
						Name:  "version",
						Value: 1,
					},
				},
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "should get invalid version error",
			fields: fields{
				postgres: nil,
			},
			args: args{
				args: []graphql.Argument{
					{
						Name:  "version",
						Value: "invalid",
					},
				},
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "should get not found version -1",
			fields: fields{
				postgres: nil,
			},
			args: args{
				args: []graphql.Argument{},
			},
			want:    -1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.GetFieldVersion(tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("DendriteService.GetFieldVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DendriteService.GetFieldVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDendriteService_GetSelectionsByField(t *testing.T) {
	type fields struct {
		backend backend.Backend
	}
	type args struct {
		field graphql.Field
		base  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []dto.Selection
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &DendriteService{
				backend: tt.fields.backend,
			}
			got, err := s.GetSelectionsByField(tt.args.field, tt.args.base)
			if (err != nil) != tt.wantErr {
				t.Errorf("DendriteService.GetSelectionsByField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DendriteService.GetSelectionsByField() = %v, want %v", got, tt.want)
			}
		})
	}
}
