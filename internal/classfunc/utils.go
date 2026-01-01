// Package classfunc provides tools for generating class-related code and queries.
package classfunc

import "fmt"

func generateFunctionName(schema string, name string) string {
	return fmt.Sprintf("%s.%s", schema, name)
}

func generateDropFunctionSQL(schema string, name string) string {
	return fmt.Sprintf("DROP FUNCTION IF EXISTS %s.%s;", schema, name)
}
