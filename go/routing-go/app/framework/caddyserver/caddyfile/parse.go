package caddyfile

// ServerBlock associates any number of keys (usually addresses
// of some sort) with tokens (grouped by directive name).
type ServerBlock struct {
	Keys   []string
	Tokens map[string][]Token
}
