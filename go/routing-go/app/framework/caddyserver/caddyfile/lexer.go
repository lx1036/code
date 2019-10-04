package caddyfile

import "bufio"

type (
	// lexer is a utility which can get values, token by
	// token, from a Reader. A token is a word, and tokens
	// are separated by whitespace. A word can be enclosed
	// in quotes if it contains whitespace.
	lexer struct {
		reader *bufio.Reader
		token  Token
		line   int
	}

	// Token represents a single parsable unit.
	Token struct {
		File string
		Line int
		Text string
	}
)
