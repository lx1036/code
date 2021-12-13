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

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HookConfiguration) DeepCopyInto(out *HookConfiguration) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	if in.WebHooks != nil {
		in, out := &in.WebHooks, &out.WebHooks
		*out = make(WebHooks, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HookConfiguration.
func (in *HookConfiguration) DeepCopy() *HookConfiguration {
	if in == nil {
		return nil
	}
	out := new(HookConfiguration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *HookConfiguration) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HookStage) DeepCopyInto(out *HookStage) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HookStage.
func (in *HookStage) DeepCopy() *HookStage {
	if in == nil {
		return nil
	}
	out := new(HookStage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in HookStageList) DeepCopyInto(out *HookStageList) {
	{
		in := &in
		*out = make(HookStageList, len(*in))
		copy(*out, *in)
		return
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HookStageList.
func (in HookStageList) DeepCopy() HookStageList {
	if in == nil {
		return nil
	}
	out := new(HookStageList)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebHook) DeepCopyInto(out *WebHook) {
	*out = *in
	if in.Stages != nil {
		in, out := &in.Stages, &out.Stages
		*out = make(HookStageList, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebHook.
func (in *WebHook) DeepCopy() *WebHook {
	if in == nil {
		return nil
	}
	out := new(WebHook)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in WebHooks) DeepCopyInto(out *WebHooks) {
	{
		in := &in
		*out = make(WebHooks, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
		return
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebHooks.
func (in WebHooks) DeepCopy() WebHooks {
	if in == nil {
		return nil
	}
	out := new(WebHooks)
	in.DeepCopyInto(out)
	return *out
}
