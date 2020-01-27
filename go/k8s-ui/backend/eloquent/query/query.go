package query

type Query struct {
	From       string
	Wheres     []*Where
	Columns    []string
	JoinClause bool
}

type Where struct {
	Type     string
	Sql      string
	Column   string
	First    string
	Second   string
	Operator string
	Value    interface{}
	Values   []interface{}
	Boolean  string
	Not      bool
}
