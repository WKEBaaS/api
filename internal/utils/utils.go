// Package utils provides helper functions for various operations.
package utils

// Helper function to convert bool to string
func BoolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
