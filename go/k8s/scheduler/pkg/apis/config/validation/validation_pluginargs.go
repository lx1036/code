package validation

import (
	"fmt"

	configv1 "k8s-lx1036/k8s/scheduler/pkg/apis/config/v1"
)

// ValidateNodeLabelArgs validates that NodeLabelArgs are correct.
func ValidateNodeLabelArgs(args configv1.NodeLabelArgs) error {
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

func ValidateNodeResourcesLeastAllocatedArgs(args *configv1.NodeResourcesLeastAllocatedArgs) error {
	return validateResources(args.Resources)
}

func validateResources(resources []configv1.ResourceSpec) error {
	for _, resource := range resources {
		if resource.Weight <= 0 {
			return fmt.Errorf("resource Weight of %v should be a positive value, got %v", resource.Name, resource.Weight)
		}
		if resource.Weight > 100 {
			return fmt.Errorf("resource Weight of %v should be less than 100, got %v", resource.Name, resource.Weight)
		}
	}
	return nil
}
