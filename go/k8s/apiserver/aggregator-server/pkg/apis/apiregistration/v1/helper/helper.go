package helper

import (
	"k8s-lx1036/k8s/apiserver/aggregator-server/pkg/apis/apiregistration/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewLocalAvailableAPIServiceCondition returns a condition for an available local APIService.
func NewLocalAvailableAPIServiceCondition() v1.APIServiceCondition {
	return v1.APIServiceCondition{
		Type:               v1.Available,
		Status:             v1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             "Local",
		Message:            "Local APIServices are always available",
	}
}

// SetAPIServiceCondition sets the status condition.  It either overwrites the existing one or
// creates a new one
func SetAPIServiceCondition(apiService *v1.APIService, newCondition v1.APIServiceCondition) {
	existingCondition := GetAPIServiceConditionByType(apiService, newCondition.Type)
	if existingCondition == nil {
		apiService.Status.Conditions = append(apiService.Status.Conditions, newCondition)
		return
	}

	if existingCondition.Status != newCondition.Status {
		existingCondition.Status = newCondition.Status
		existingCondition.LastTransitionTime = newCondition.LastTransitionTime
	}

	existingCondition.Reason = newCondition.Reason
	existingCondition.Message = newCondition.Message
}

// GetAPIServiceConditionByType gets an *APIServiceCondition by APIServiceConditionType if present
func GetAPIServiceConditionByType(apiService *v1.APIService, conditionType v1.APIServiceConditionType) *v1.APIServiceCondition {
	for i := range apiService.Status.Conditions {
		if apiService.Status.Conditions[i].Type == conditionType {
			return &apiService.Status.Conditions[i]
		}
	}
	return nil
}

// IsAPIServiceConditionTrue indicates if the condition is present and strictly true
func IsAPIServiceConditionTrue(apiService *v1.APIService, conditionType v1.APIServiceConditionType) bool {
	condition := GetAPIServiceConditionByType(apiService, conditionType)
	return condition != nil && condition.Status == v1.ConditionTrue
}
