// Package dot provides general DNS-over-TLS functionality.
package dot

// IsValidResponse returns a boolean indicating if the []string carries a response.
func IsValidResponse(r []string) bool {
	return len(r) > 0
}
