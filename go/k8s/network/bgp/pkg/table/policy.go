package table

import "sync"

type RoutingPolicy struct {
	mu sync.RWMutex

	definedSetMap DefinedSetMap
	policyMap     map[string]*Policy
	statementMap  map[string]*Statement
	assignmentMap map[string]*Assignment
}

func NewRoutingPolicy() *RoutingPolicy {
	return &RoutingPolicy{
		definedSetMap: make(map[DefinedType]map[string]DefinedSet),
		policyMap:     make(map[string]*Policy),
		statementMap:  make(map[string]*Statement),
		assignmentMap: make(map[string]*Assignment),
	}
}
