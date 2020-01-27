package grammar

import (
	"k8s-lx1036/k8s-ui/backend/eloquent/query"
	"strings"
)

type Grammar interface {
	CompileSelect(*query.Query) string
}

type MysqlGrammar struct {
}

// compile *query.Query to sql string
func (grammar *MysqlGrammar) CompileSelect(query *query.Query) string {
	
	sql := strings.TrimSpace(grammar.concatenate(grammar.compileComponent(query)))
	
	return sql
}

func (grammar *MysqlGrammar) concatenate(component interface{}) string {
	
}

func (grammar *MysqlGrammar) compileComponent(query2 *query.Query) []string {
	
}
