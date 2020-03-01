package api

// ResourceKind is an unique name for each resource. It can used for API discovery and generic
// code that does things based on the kind. For example, there may be a generic "deleter"
// that based on resource kind, name and namespace deletes it.
type ResourceKind string


// ListMeta describes list of objects, i.e. holds information about pagination options set for the list.
type ListMeta struct {
	// Total number of items on the list. Used for pagination.
	TotalItems int `json:"totalItems"`
}
