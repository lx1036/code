package eloquent

import (
	"k8s-lx1036/k8s-ui/backend/eloquent/grammar"
	"k8s-lx1036/k8s-ui/backend/eloquent/query"
)

type Builder struct {
	Connection *Connection
	grammar    grammar.Grammar

	Query *query.Query
}

func NewBuilder(connection *Connection, grammar grammar.Grammar) *Builder {
	return &Builder{
		Connection: connection,
		grammar:    grammar,
	}
}

func (builder *Builder) From(table string) *Builder {
	builder.Query.From = table
	return builder
}
