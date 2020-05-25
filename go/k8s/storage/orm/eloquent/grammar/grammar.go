package grammar

import (
	"strings"
)

var selectComponents = []string{
	"aggregate",
	"columns",
	"from",
	"joins",
	"wheres",
	"groups",
	"havings",
	"orders",
	"limit",
	"offset",
	// "unions",
	// "lock",
}

type Grammar interface {
	CompileSelect(*query.Query) string
}

type MysqlGrammar struct {
	Grammar
	tablePrefix string
}

// compile *query.Query to sql string
func (grammar *MysqlGrammar) CompileSelect(query *query.Query) string {

	sql := strings.TrimSpace(grammar.concatenate(grammar.compileComponent(query)))

	return sql
}

func (grammar *MysqlGrammar) concatenate(segments []string) string {
	s := ""
	for _, segment := range segments {
		if len(segment) == 0 {
			continue
		}
		if len(s) > 0 {
			s = s + " "
		}
		s = s + segment
	}
	return s
}

func (grammar *MysqlGrammar) compileComponent(query *query.Query) []string {
	var sql []string
	for _, component := range selectComponents {
		switch component {
		case "from":
			if len(query.From) > 0 {
				sql = append(sql, grammar.compileFrom(query, query.From))
			}
		case "columns":
			if len(query.Columns) > 0 {
				sql = append(sql, grammar.compileColumns(query, query.Columns))
			}
		case "wheres":
			if len(query.Wheres) > 0 {
				sql = append(sql, grammar.compileWheres(query, query.Wheres))
			}
		}
	}

	return sql
}

func (grammar *MysqlGrammar) compileFrom(query2 *query.Query, table string) string {
	return "FROM " + grammar.WrapTable(table)
}

func (grammar *MysqlGrammar) WrapTable(table string) string {
	return grammar.tablePrefix + table
}

func (grammar *MysqlGrammar) compileColumns(query *query.Query, columns []string) string {
	sel := "SELECT "

	return sel + strings.Join(columns, ", ")
}

func (grammar *MysqlGrammar) compileWheres(query *query.Query, wheres []*query.Where) string {
	var sql []string
	for _, where := range wheres {
		w := ""

		switch where.Type {
		case "Basic":
			w = where.Boolean + " " + grammar.whereBasic(query, where)
		case "In":
		}

		sql = append(sql, w)
	}

	if len(sql) > 0 {
		conjunction := "WHERE"
		if query.JoinClause {
			conjunction = "ON"
		}
		return conjunction + " " + removeLeadingBoolean(strings.Join(sql, " "))
	}

	return ""
}

func removeLeadingBoolean(join string) string {
	return join
}

func (grammar *MysqlGrammar) whereBasic(query2 *query.Query, where *query.Where) string {
	return grammar.Wrap(where.Column, false) + " " + where.Operator + " ?"
}

func (grammar *MysqlGrammar) Wrap(value string, prefixAlias bool) string {
	return grammar.wrapSegments(strings.Split(value, "."))
}

func (grammar *MysqlGrammar) wrapSegments(segments []string) string {
	for key, segment := range segments {
		if key == 0 && len(segments) > 1 {
			segments[key] = grammar.WrapTable(segment)
		} else {
			segments[key] = grammar.wapValue(segment)
		}
	}
	return strings.Join(segments, ".")
}

func (grammar *MysqlGrammar) wapValue(value string) string {
	if value != "*" {
		return "`" + strings.Replace(value, "`", "``", -1) + "`"
	}
	return value
}
