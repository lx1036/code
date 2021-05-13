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

// Code generated by defaulter-gen. DO NOT EDIT.

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// RegisterDefaults adds defaulters functions to the given scheme.
// Public to allow building arbitrary schemes.
// All generated defaulters are covering - they call all nested defaulters.
func RegisterDefaults(scheme *runtime.Scheme) error {
	scheme.AddTypeDefaultingFunc(&CustomResourceDefinition{}, func(obj interface{}) { SetObjectDefaults_CustomResourceDefinition(obj.(*CustomResourceDefinition)) })
	scheme.AddTypeDefaultingFunc(&CustomResourceDefinitionList{}, func(obj interface{}) {
		SetObjectDefaults_CustomResourceDefinitionList(obj.(*CustomResourceDefinitionList))
	})
	return nil
}

func SetObjectDefaults_CustomResourceDefinition(in *CustomResourceDefinition) {
	SetDefaults_CustomResourceDefinition(in)
	SetDefaults_CustomResourceDefinitionSpec(&in.Spec)
	if in.Spec.Conversion != nil {
		if in.Spec.Conversion.Webhook != nil {
			if in.Spec.Conversion.Webhook.ClientConfig != nil {
				if in.Spec.Conversion.Webhook.ClientConfig.Service != nil {
					SetDefaults_ServiceReference(in.Spec.Conversion.Webhook.ClientConfig.Service)
				}
			}
		}
	}
}

func SetObjectDefaults_CustomResourceDefinitionList(in *CustomResourceDefinitionList) {
	for i := range in.Items {
		a := &in.Items[i]
		SetObjectDefaults_CustomResourceDefinition(a)
	}
}
