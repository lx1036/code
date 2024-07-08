package labels

type OpLabels struct {
	// Active labels that are enabled and disabled but not deleted
	Custom Labels

	// Labels derived from orchestration system
	OrchestrationIdentity Labels

	// orchestrationIdentity labels which have been disabled
	Disabled Labels

	// orchestrationInfo - labels from orchestration which are not used in determining a security identity
	OrchestrationInfo Labels
}

// NewOpLabels creates new initialized OpLabels
func NewOpLabels() OpLabels {
	return OpLabels{
		Custom:                Labels{},
		Disabled:              Labels{},
		OrchestrationIdentity: Labels{},
		OrchestrationInfo:     Labels{},
	}
}

// AllLabels returns all Labels within the provided OpLabels.
func (o *OpLabels) AllLabels() Labels {
	all := make(Labels, len(o.Custom)+len(o.OrchestrationInfo)+len(o.OrchestrationIdentity)+len(o.Disabled))

	for k, v := range o.Custom {
		all[k] = v
	}

	for k, v := range o.Disabled {
		all[k] = v
	}

	for k, v := range o.OrchestrationIdentity {
		all[k] = v
	}

	for k, v := range o.OrchestrationInfo {
		all[k] = v
	}
	return all
}
