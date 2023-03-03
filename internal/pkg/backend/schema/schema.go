package schema

import _ "embed"

//go:embed config.sql
var schema string

func GetSchema() string {
	return schema
}
