package cm

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"strings"
	"testing"
)

// INFO: 测试 pod level cgroups /sys/fs/cgroup/cpuset/kubepods/burstable/pod2631ec27-8cbe-45f2-984b-a61e866ab4a2
func TestIsCgroupPod(test *testing.T) {
	qosContainersInfo := QOSContainersInfo{
		Guaranteed: RootCgroupName,
		Burstable:  NewCgroupName(RootCgroupName, strings.ToLower(string(v1.PodQOSBurstable))),
		BestEffort: NewCgroupName(RootCgroupName, strings.ToLower(string(v1.PodQOSBestEffort))),
	}
	podUID := types.UID("8dbc9e11-dc93-4eb2-bbef-ed21cf1db420")
	testCases := []struct {
		description    string
		input          CgroupName
		expectedResult bool
		expectedUID    types.UID
	}{
		{
			description:    "test",
			input:          RootCgroupName,
			expectedResult: false,
			expectedUID:    types.UID(""),
		},
		{
			input:          NewCgroupName(qosContainersInfo.BestEffort, GetPodCgroupNameSuffix(podUID)),
			expectedResult: true,
			expectedUID:    podUID,
		},
	}
	for _, cgroupDriver := range []string{string(libcontainerCgroupfs), string(libcontainerSystemd)} {
		podContainerManager := &podContainerManagerImpl{
			cgroupManager:     NewCgroupManager(nil, cgroupDriver),
			enforceCPULimits:  true,
			qosContainersInfo: qosContainersInfo,
		}

		for _, testCase := range testCases {
			test.Run(testCase.description, func(t *testing.T) {
				// give the right cgroup structure based on driver
				cgroupfs := testCase.input.ToCgroupfs()
				if cgroupDriver == "systemd" {
					cgroupfs = testCase.input.ToSystemd()
				}

				result, resultUID := podContainerManager.IsPodCgroup(cgroupfs)
				if result != testCase.expectedResult {
					t.Errorf("Unexpected result for driver: %v, input: %v, expected: %v, actual: %v", cgroupDriver, testCase.input, testCase.expectedResult, result)
				}
				if resultUID != testCase.expectedUID {
					t.Errorf("Unexpected result for driver: %v, input: %v, expected: %v, actual: %v", cgroupDriver, testCase.input, testCase.expectedUID, resultUID)
				}
			})
		}
	}

}
