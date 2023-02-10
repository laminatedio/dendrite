package dendrite

import (
	"context"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laminatedio/dendrite/internal/pkg/backend"
	"github.com/laminatedio/dendrite/internal/pkg/dendrite/dto"
	backendmock "github.com/laminatedio/dendrite/mocks/internal_/pkg/backend"
	"github.com/tmc/graphql"
)

type Config struct {
	Path   string
	Values []string
}

type MetaData struct {
	Path           string
	LatestVersion  int
	CurrentVersion int
}

var postgresDsn = os.Getenv("POSTGRES_DSN")

var testConfig = []Config{
	{
		Path:   "/A/B/C",
		Values: []string{"1"},
	},
	{
		Path:   "/A/B/D",
		Values: []string{"2"},
	},
	{
		Path:   "/A/B",
		Values: []string{"C", "D"},
	},
	{
		Path:   "/E",
		Values: []string{"3"},
	},
}

func ImportTestData(ctx context.Context, configs []Config, worker backend.Backend) error {
	for _, config := range configs {
		_, err := worker.SetMany(ctx, config.Path, config.Values, backend.SetOptions{KeepCurrent: false})
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
var mock = &backendmock.Backend{}
var mockS = DendriteService{
	backend: mock,
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
			name:    "should get data by query",
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
			name:    "should get data by query (with one level)",
			configs: testConfig,
			args: args{
				ctx: context.Background(),
				query: `{
					E
				}`,
			},
			want: map[string]any{
				"E": "3",
			},
			wantErr: false,
		},
		{
			name: "should get version 2",
			configs: []Config{
				{
					Path:   "/A",
					Values: []string{"1"},
				},
				{
					Path:   "/A",
					Values: []string{"2"},
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
			name:    "should have some key not found",
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
			name:    "should be invalid query",
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
	type args struct {
		args []graphql.Argument
	}
	tests := []struct {
		name        string
		args        args
		testingData map[string]string
		want        int
		wantErr     bool
	}{
		// TODO: Add test cases.
		{
			name: "should get version in args",
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
			args: args{
				args: []graphql.Argument{},
			},
			want:    -1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mockS.GetFieldVersion(tt.args.args)
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
	type args struct {
		field graphql.Field
		base  string
	}
	tests := []struct {
		name    string
		args    args
		want    []dto.Selection
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "should get the selections by field",
			args: args{
				field: graphql.Field{
					Name: "A",
					SelectionSet: graphql.SelectionSet{
						{
							Field: &graphql.Field{
								Name: "B",
								Arguments: graphql.Arguments{
									{
										Name:  "version",
										Value: 1,
									},
								},
							},
						},
						{
							Field: &graphql.Field{
								Name: "C",
								SelectionSet: graphql.SelectionSet{
									{
										Field: &graphql.Field{
											Name: "D",
											Arguments: graphql.Arguments{
												{
													Name:  "version",
													Value: 2,
												},
											},
										},
									},
									{
										Field: &graphql.Field{
											Name: "E",
										},
									},
								},
							},
						},
					},
				},
			},
			want: []dto.Selection{
				{
					Path:    "/A/B",
					Version: 1,
				},
				{
					Path:    "/A/C/D",
					Version: 2,
				},
				{
					Path:    "/A/C/E",
					Version: -1,
				},
			},
		},
		{
			name: "should return error (invalid version)",
			args: args{
				field: graphql.Field{
					Name: "A",
					Arguments: graphql.Arguments{
						{
							Name:  "version",
							Value: "testing",
						},
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mockS.GetSelectionsByField(tt.args.field, tt.args.base)
			if (err != nil) != tt.wantErr {
				t.Errorf("DendriteService.GetSelectionsByField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DendriteService.GetSelectionsByField() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestDendriteService_GetObjectByPaths(t *testing.T) {
	type args struct {
		ctx        context.Context
		selections []dto.Selection
	}
	mock.On("GetMany", context.Background(), "/A/B/C", 1).Return([]string{"1", "2"}, nil)
	mock.On("GetMany", context.Background(), "/A/B/D", 2).Return([]string{"3"}, nil)
	mock.On("GetManyCurrent", context.Background(), "/A/B").Return([]string{"C", "D"}, nil)
	ctx := context.Background()
	tests := []struct {
		name    string
		args    args
		want    map[string]any
		wantErr bool
	}{
		{
			name: "should return the desired object",
			args: args{
				ctx: ctx,
				selections: []dto.Selection{
					{
						Path:    "/A/B/C",
						Version: 1,
					},
					{
						Path:    "/A/B/D",
						Version: 2,
					},
					{
						Path:    "/A/B",
						Version: -1,
					},
				},
			},
			want: map[string]any{
				"A": map[string]any{
					"B": map[string]any{
						`/`: []string{"C", "D"},
						"C": []string{"1", "2"},
						"D": "3",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "want err :invalid path",
			args: args{
				ctx: ctx,
				selections: []dto.Selection{
					{
						Path:    "A/B/C",
						Version: 1,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mockS.GetObjectByPaths(tt.args.ctx, tt.args.selections)
			if (err != nil) != tt.wantErr {
				t.Errorf("DendriteService.GetObjectByPaths() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DendriteService.GetObjectByPaths() = %v, want %v", got, tt.want)
			}
		})
	}
}
