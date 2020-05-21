package eloquent

import (
	"reflect"
)

type Builder struct {
	Connection *Connection
	grammar    grammar.Grammar

	Query    *query.Query
	Bindings map[string][]interface{}
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

func (builder *Builder) Where(column string, args ...interface{}) *Builder {
	boolean := "and"
	count := len(args)

	var operator string
	var value interface{}

	if count == 1 {
		operator = "="
		value = args[0]
	}

	where := &query.Where{
		Type:     "Basic",
		Column:   column,
		Operator: operator,
		Value:    value,
		Boolean:  boolean,
	}

	builder.Query.Wheres = append(builder.Query.Wheres, where)

	builder.AddBinding(value, "where")

	return builder
}

func (builder *Builder) AddBinding(value interface{}, segment string) {
	if _, ok := builder.Bindings[segment]; !ok {
		return
	}

	t := reflect.TypeOf(value)

	switch t.Kind() {
	case reflect.Slice:

	default:
		builder.Bindings[segment] = append(builder.Bindings[segment], value)
	}
}

func (builder *Builder) Get(dest interface{}) error {

	err := builder.runSelect(dest)

	return err
}

func (builder *Builder) runSelect(dest interface{}) error {
	return builder.Connection.Select(builder.ToSql(), builder.GetBindings(), dest)
}

func (builder *Builder) ToSql() string {
	return builder.grammar.CompileSelect(builder.Query)
}

func (builder *Builder) GetBindings() []interface{} {
	var bindings []interface{}
	for _, val := range builder.Bindings {
		for _, v := range val {
			bindings = append(bindings, v)
		}
	}
	return bindings
}
