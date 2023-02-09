package backend

import (
	"context"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresBackend_Set(t *testing.T) {
	type fields struct {
		Conn *pgxpool.Pool
	}
	type args struct {
		ctx     context.Context
		path    string
		value   string
		options SetOptions
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Metadata
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &PostgresBackend{
				Conn: tt.fields.Conn,
			}
			got, err := b.Set(tt.args.ctx, tt.args.path, tt.args.value, tt.args.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("PostgresBackend.Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PostgresBackend.Set() = %v, want %v", got, tt.want)
			}
		})
	}
}
