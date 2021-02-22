package main

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

// pluginInfoFormat is the default Go template for emitting a
// csi.GetPluginInfoResponse
const pluginInfoFormat = `{{printf "%q\t%q" .Name .VendorVersion}}` +
	`{{range $k, $v := .Manifest}}{{printf "\t%q=%q" $k $v}}{{end}}` +
	`{{"\n"}}`

// pluginCapsFormat is the default Go template for emitting a
// csi.GetPluginCapabilities
const pluginCapsFormat = `{{range $v := .Capabilities}}` +
	`{{with $t := .Type}}` +
	`{{if isa $t "*csi.PluginCapability_Service_"}}{{if $t.Service}}` +
	`{{printf "%s\n" $t.Service.Type}}` +
	`{{end}}{{end}}` +
	`{{if isa $t "*csi.PluginCapability_VolumeExpansion_"}}{{if $t.VolumeExpansion}}` +
	`{{printf "%s\n" $t.VolumeExpansion.Type}}` +
	`{{end}}{{end}}` +
	`{{end}}` +
	`{{end}}`

var identity struct {
	client csi.IdentityClient
}

var identityCmd = &cobra.Command{
	Use: "identity",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if f := cmd.Root().PersistentPreRunE; f != nil {
			if err := f(cmd, args); err != nil {
				return err
			}
		}

		identity.client = csi.NewIdentityClient(ClientConn)

		return nil
	},
}

// debug: go run . identity plugin-info --endpoint 127.0.0.1:10000
var pluginInfoCmd = &cobra.Command{
	Use: "plugin-info",
	RunE: func(cmd *cobra.Command, args []string) error {
		response, err := identity.client.GetPluginInfo(context.TODO(), &csi.GetPluginInfoRequest{})
		if err != nil {
			return err
		}

		klog.Infof("GetPluginInfo response: %v", response)

		return nil
	},
}

// debug: go run . identity plugin-capability --endpoint 127.0.0.1:10000
var pluginCapabilityCmd = &cobra.Command{
	Use: "plugin-capability",
	RunE: func(cmd *cobra.Command, args []string) error {
		response, err := identity.client.GetPluginCapabilities(context.TODO(), &csi.GetPluginCapabilitiesRequest{})
		if err != nil {
			return err
		}

		klog.Infof("%v", response)

		return nil
	},
}

func init() {
	RootCmd.AddCommand(identityCmd)
	identityCmd.AddCommand(pluginInfoCmd)
	identityCmd.AddCommand(pluginCapabilityCmd)
}
