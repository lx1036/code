//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by conversion-gen. DO NOT EDIT.

package v1alpha1

import (
	componentconfig "k8s-lx1036/k8s/scheduler/descheduler/pkg/apis/componentconfig"
	time "time"

	conversion "k8s.io/apimachinery/pkg/conversion"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

func init() {
	localSchemeBuilder.Register(RegisterConversions)
}

// RegisterConversions adds conversion functions to the given scheme.
// Public to allow building arbitrary schemes.
func RegisterConversions(s *runtime.Scheme) error {
	if err := s.AddGeneratedConversionFunc((*DeschedulerConfiguration)(nil), (*componentconfig.DeschedulerConfiguration)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_DeschedulerConfiguration_To_componentconfig_DeschedulerConfiguration(a.(*DeschedulerConfiguration), b.(*componentconfig.DeschedulerConfiguration), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*componentconfig.DeschedulerConfiguration)(nil), (*DeschedulerConfiguration)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_componentconfig_DeschedulerConfiguration_To_v1alpha1_DeschedulerConfiguration(a.(*componentconfig.DeschedulerConfiguration), b.(*DeschedulerConfiguration), scope)
	}); err != nil {
		return err
	}
	return nil
}

func autoConvert_v1alpha1_DeschedulerConfiguration_To_componentconfig_DeschedulerConfiguration(in *DeschedulerConfiguration, out *componentconfig.DeschedulerConfiguration, s conversion.Scope) error {
	out.DeschedulingInterval = time.Duration(in.DeschedulingInterval)
	out.KubeconfigFile = in.KubeconfigFile
	out.PolicyConfigFile = in.PolicyConfigFile
	out.DryRun = in.DryRun
	out.NodeSelector = in.NodeSelector
	out.MaxNoOfPodsToEvictPerNode = in.MaxNoOfPodsToEvictPerNode
	out.EvictLocalStoragePods = in.EvictLocalStoragePods
	out.IgnorePVCPods = in.IgnorePVCPods
	out.Logging = in.Logging
	return nil
}

// Convert_v1alpha1_DeschedulerConfiguration_To_componentconfig_DeschedulerConfiguration is an autogenerated conversion function.
func Convert_v1alpha1_DeschedulerConfiguration_To_componentconfig_DeschedulerConfiguration(in *DeschedulerConfiguration, out *componentconfig.DeschedulerConfiguration, s conversion.Scope) error {
	return autoConvert_v1alpha1_DeschedulerConfiguration_To_componentconfig_DeschedulerConfiguration(in, out, s)
}

func autoConvert_componentconfig_DeschedulerConfiguration_To_v1alpha1_DeschedulerConfiguration(in *componentconfig.DeschedulerConfiguration, out *DeschedulerConfiguration, s conversion.Scope) error {
	out.DeschedulingInterval = time.Duration(in.DeschedulingInterval)
	out.KubeconfigFile = in.KubeconfigFile
	out.PolicyConfigFile = in.PolicyConfigFile
	out.DryRun = in.DryRun
	out.NodeSelector = in.NodeSelector
	out.MaxNoOfPodsToEvictPerNode = in.MaxNoOfPodsToEvictPerNode
	out.EvictLocalStoragePods = in.EvictLocalStoragePods
	out.IgnorePVCPods = in.IgnorePVCPods
	out.Logging = in.Logging
	return nil
}

// Convert_componentconfig_DeschedulerConfiguration_To_v1alpha1_DeschedulerConfiguration is an autogenerated conversion function.
func Convert_componentconfig_DeschedulerConfiguration_To_v1alpha1_DeschedulerConfiguration(in *componentconfig.DeschedulerConfiguration, out *DeschedulerConfiguration, s conversion.Scope) error {
	return autoConvert_componentconfig_DeschedulerConfiguration_To_v1alpha1_DeschedulerConfiguration(in, out, s)
}
