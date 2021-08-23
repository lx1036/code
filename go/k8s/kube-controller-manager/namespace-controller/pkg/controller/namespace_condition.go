package controller

import (
	"fmt"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
)

type namespaceConditionUpdater struct {
	newConditions       []corev1.NamespaceCondition
	deleteContentErrors []error
}

// ProcessDiscoverResourcesErr creates error condition from ErrGroupDiscoveryFailed.
func (u *namespaceConditionUpdater) ProcessDiscoverResourcesErr(err error) {
	var msg string
	if derr, ok := err.(*discovery.ErrGroupDiscoveryFailed); ok {
		msg = fmt.Sprintf("Discovery failed for some groups, %d failing: %v", len(derr.Groups), err)
	} else {
		msg = err.Error()
	}

	d := corev1.NamespaceCondition{
		Type:               corev1.NamespaceDeletionDiscoveryFailure,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             "DiscoveryFailed",
		Message:            msg,
	}

	u.newConditions = append(u.newConditions, d)
}

// ProcessGroupVersionErr creates error condition if parsing GroupVersion of resources fails.
func (u *namespaceConditionUpdater) ProcessGroupVersionErr(err error) {
	d := corev1.NamespaceCondition{
		Type:               corev1.NamespaceDeletionGVParsingFailure,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             "GroupVersionParsingFailed",
		Message:            err.Error(),
	}

	u.newConditions = append(u.newConditions, d)
}

// ProcessDeleteContentErr creates error condition from multiple delete content errors.
func (u *namespaceConditionUpdater) ProcessDeleteContentErr(err error) {
	u.deleteContentErrors = append(u.deleteContentErrors, err)
}

// ProcessContentTotals may create conditions for NamespaceContentRemaining and NamespaceFinalizersRemaining.
func (u *namespaceConditionUpdater) ProcessContentTotals(contentTotals allGVRDeletionMetadata) {
	if len(contentTotals.gvrToNumRemaining) != 0 {
		var remainingResources []string
		for gvr, numRemaining := range contentTotals.gvrToNumRemaining {
			if numRemaining == 0 {
				continue
			}
			remainingResources = append(remainingResources, fmt.Sprintf("%s.%s has %d resource instances", gvr.Resource, gvr.Group, numRemaining))
		}
		// sort for stable updates
		sort.Strings(remainingResources)
		u.newConditions = append(u.newConditions, corev1.NamespaceCondition{
			Type:               corev1.NamespaceContentRemaining,
			Status:             corev1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
			Reason:             "SomeResourcesRemain",
			Message:            fmt.Sprintf("Some resources are remaining: %s", strings.Join(remainingResources, ", ")),
		})
	}

	if len(contentTotals.finalizersToNumRemaining) != 0 {
		var remainingByFinalizer []string
		for finalizer, numRemaining := range contentTotals.finalizersToNumRemaining {
			if numRemaining == 0 {
				continue
			}
			remainingByFinalizer = append(remainingByFinalizer, fmt.Sprintf("%s in %d resource instances", finalizer, numRemaining))
		}
		// sort for stable updates
		sort.Strings(remainingByFinalizer)
		u.newConditions = append(u.newConditions, corev1.NamespaceCondition{
			Type:               corev1.NamespaceFinalizersRemaining,
			Status:             corev1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
			Reason:             "SomeFinalizersRemain",
			Message:            fmt.Sprintf("Some content in the namespace has finalizers remaining: %s", strings.Join(remainingByFinalizer, ", ")),
		})
	}
}

func makeDeleteContentCondition(err []error) *corev1.NamespaceCondition {
	if len(err) == 0 {
		return nil
	}
	msgs := make([]string, 0, len(err))
	for _, e := range err {
		msgs = append(msgs, e.Error())
	}
	sort.Strings(msgs)
	return &corev1.NamespaceCondition{
		Type:               corev1.NamespaceDeletionContentFailure,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             "ContentDeletionFailed",
		Message:            fmt.Sprintf("Failed to delete all resource types, %d remaining: %v", len(err), strings.Join(msgs, ", ")),
	}
}

func getCondition(conditions []corev1.NamespaceCondition, conditionType corev1.NamespaceConditionType) *corev1.NamespaceCondition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return &(conditions[i])
		}
	}
	return nil
}

var (
	// conditionTypes Namespace condition types that are maintained by namespace_deleter controller.
	conditionTypes = []corev1.NamespaceConditionType{
		corev1.NamespaceDeletionDiscoveryFailure,
		corev1.NamespaceDeletionGVParsingFailure,
		corev1.NamespaceDeletionContentFailure,
		corev1.NamespaceContentRemaining,
		corev1.NamespaceFinalizersRemaining,
	}
	okMessages = map[corev1.NamespaceConditionType]string{
		corev1.NamespaceDeletionDiscoveryFailure: "All resources successfully discovered",
		corev1.NamespaceDeletionGVParsingFailure: "All legacy kube types successfully parsed",
		corev1.NamespaceDeletionContentFailure:   "All content successfully deleted, may be waiting on finalization",
		corev1.NamespaceContentRemaining:         "All content successfully removed",
		corev1.NamespaceFinalizersRemaining:      "All content-preserving finalizers finished",
	}
	okReasons = map[corev1.NamespaceConditionType]string{
		corev1.NamespaceDeletionDiscoveryFailure: "ResourcesDiscovered",
		corev1.NamespaceDeletionGVParsingFailure: "ParsedGroupVersions",
		corev1.NamespaceDeletionContentFailure:   "ContentDeleted",
		corev1.NamespaceContentRemaining:         "ContentRemoved",
		corev1.NamespaceFinalizersRemaining:      "ContentHasNoFinalizers",
	}
)

func newSuccessfulCondition(conditionType corev1.NamespaceConditionType) *corev1.NamespaceCondition {
	return &corev1.NamespaceCondition{
		Type:               conditionType,
		Status:             corev1.ConditionFalse,
		LastTransitionTime: metav1.Now(),
		Reason:             okReasons[conditionType],
		Message:            okMessages[conditionType],
	}
}

func updateConditions(status *corev1.NamespaceStatus, newConditions []corev1.NamespaceCondition) (hasChanged bool) {
	for _, conditionType := range conditionTypes {
		newCondition := getCondition(newConditions, conditionType)
		// if we weren't failing, then this returned nil.  We should set the "ok" variant of the condition
		if newCondition == nil {
			newCondition = newSuccessfulCondition(conditionType)
		}
		oldCondition := getCondition(status.Conditions, conditionType)

		// only new condition of this type exists, add to the list
		if oldCondition == nil {
			status.Conditions = append(status.Conditions, *newCondition)
			hasChanged = true
		} else if oldCondition.Status != newCondition.Status || oldCondition.Message != newCondition.Message || oldCondition.Reason != newCondition.Reason {
			// old condition needs to be updated
			if oldCondition.Status != newCondition.Status {
				oldCondition.LastTransitionTime = metav1.Now()
			}
			oldCondition.Type = newCondition.Type
			oldCondition.Status = newCondition.Status
			oldCondition.Reason = newCondition.Reason
			oldCondition.Message = newCondition.Message
			hasChanged = true
		}
	}
	return
}

// Update compiles processed errors from namespace deletion into status conditions.
func (u *namespaceConditionUpdater) Update(ns *corev1.Namespace) bool {
	if c := getCondition(u.newConditions, corev1.NamespaceDeletionContentFailure); c == nil {
		if c := makeDeleteContentCondition(u.deleteContentErrors); c != nil {
			u.newConditions = append(u.newConditions, *c)
		}
	}

	return updateConditions(&ns.Status, u.newConditions)
}
