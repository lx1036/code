package grammar

type Grammar interface {
	CompileSelect() string
}

type MysqlGrammar struct {
}

func (grammar *MysqlGrammar) CompileSelect() string {
	return ""
}
