package common

type NamespaceQuery struct {
	namespaces []string
}

func NewNamespaceQuery(namespaces []string) *NamespaceQuery {
	return &NamespaceQuery{
		namespaces: namespaces,
	}
}
