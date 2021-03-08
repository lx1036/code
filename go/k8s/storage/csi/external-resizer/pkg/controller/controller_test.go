package controller

import (
	"encoding/json"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetPVCPatchData(t *testing.T) {
	type Fixture struct {
		OldPVC *v1.PersistentVolumeClaim
	}

	for i, fixture := range []Fixture{
		{&v1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{ResourceVersion: "1"}}},
		{&v1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{ResourceVersion: "2"}}},
	} {
		newPVC := fixture.OldPVC.DeepCopy()
		newPVC.Status.Conditions = append(newPVC.Status.Conditions, v1.PersistentVolumeClaimCondition{Type: v1.PersistentVolumeClaimResizing, Status: v1.ConditionTrue})
		patchBytes, err := GetPVCPatchData(fixture.OldPVC, newPVC)
		if err != nil {
			t.Errorf("Case %d: Get patch data failed: %v", i, err)
		}

		var patchMap map[string]interface{}
		err = json.Unmarshal(patchBytes, &patchMap)
		if err != nil {
			t.Errorf("Case %d: unmarshalling json patch failed: %v", i, err)
		}

		metadata, exist := patchMap["metadata"].(map[string]interface{})
		if !exist {
			t.Errorf("Case %d: ResourceVersion should exist in patch data", i)
		}
		resourceVersion := metadata["resourceVersion"].(string)
		if resourceVersion != fixture.OldPVC.ResourceVersion {
			t.Errorf("Case %d: ResourceVersion should be %s, got %s",
				i, fixture.OldPVC.ResourceVersion, resourceVersion)
		}
	}
}
