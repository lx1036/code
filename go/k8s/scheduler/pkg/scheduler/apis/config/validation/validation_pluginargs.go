package validation

import (
	"fmt"

	"k8s-lx1036/k8s/scheduler/pkg/scheduler/apis/config"
)

// ValidateNodeLabelArgs validates that NodeLabelArgs are correct.
func ValidateNodeLabelArgs(args config.NodeLabelArgs) error {
	if err := validateNoConflict(args.PresentLabels, args.AbsentLabels); err != nil {
		return err
	}
	if err := validateNoConflict(args.PresentLabelsPreference, args.AbsentLabelsPreference); err != nil {
		return err
	}
	return nil
}

// validateNoConflict validates that presentLabels and absentLabels do not conflict.
func validateNoConflict(presentLabels []string, absentLabels []string) error {
	m := make(map[string]struct{}, len(presentLabels))
	for _, l := range presentLabels {
		m[l] = struct{}{}
	}
	for _, l := range absentLabels {
		if _, ok := m[l]; ok {
			return fmt.Errorf(`detecting at least one label (e.g., %q) that exist in both the present(%+v) and
				absent(%+v) label list`, l, presentLabels, absentLabels)
		}
	}
	return nil
}
