// Package pggen provides tools for generating PostgreSQL-related code and queries.
package pggen

import "fmt"

func generateFunctionName(schema string, name string) string {
	return fmt.Sprintf("%s.%s", schema, name)
}

func generateDropFunctionSQL(schema string, name string) string {
	return fmt.Sprintf("DROP FUNCTION IF EXISTS %s.%s;", schema, name)
}
