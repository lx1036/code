package main

import "github.com/containernetworking/cni/pkg/version"

// @see https://github.com/containernetworking/cni/blob/master/SPEC.md
var specVersionSupported = version.PluginSupports("1.0.0")

func GetSpecVersionSupported() version.PluginInfo {
	return specVersionSupported
}
