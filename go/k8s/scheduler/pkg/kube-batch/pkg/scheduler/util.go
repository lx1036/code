package scheduler

import (
	"fmt"
	"io/ioutil"
	"strings"

	"k8s-lx1036/k8s/scheduler/pkg/kube-batch/pkg/scheduler/conf"
	"k8s-lx1036/k8s/scheduler/pkg/kube-batch/pkg/scheduler/framework"

	yaml "gopkg.in/yaml.v2"
)

var defaultSchedulerConf = `
actions: "allocate, backfill"
tiers:
- plugins:
  - name: priority
  - name: gang
- plugins:
  - name: drf
  - name: predicates
  - name: proportion
  - name: nodeorder
`

func readSchedulerConf(confPath string) (string, error) {
	dat, err := ioutil.ReadFile(confPath)
	if err != nil {
		return "", err
	}
	return string(dat), nil
}

func loadSchedulerConf(confStr string) ([]framework.Action, []conf.Tier, error) {
	var actions []framework.Action

	schedulerConf := &conf.SchedulerConfiguration{}

	buf := make([]byte, len(confStr))
	copy(buf, confStr)

	if err := yaml.Unmarshal(buf, schedulerConf); err != nil {
		return nil, nil, err
	}

	// Set default settings for each plugin if not set
	for i, tier := range schedulerConf.Tiers {
		for j := range tier.Plugins {
			plugins.ApplyPluginConfDefaults(&schedulerConf.Tiers[i].Plugins[j])
		}
	}

	actionNames := strings.Split(schedulerConf.Actions, ",")
	for _, actionName := range actionNames {
		if action, found := framework.GetAction(strings.TrimSpace(actionName)); found {
			actions = append(actions, action)
		} else {
			return nil, nil, fmt.Errorf("failed to found Action %s, ignore it", actionName)
		}
	}

	return actions, schedulerConf.Tiers, nil
}
