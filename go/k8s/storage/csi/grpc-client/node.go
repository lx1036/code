package main

import (
	"context"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

var node struct {
	client csi.NodeClient
}

// nodeCmd represents the node command
var nodeCmd = &cobra.Command{
	Use:     "node",
	Aliases: []string{"n"},
	Short:   "the csi node service rpcs",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if f := cmd.Root().PersistentPreRunE; f != nil {
			if err := f(cmd, args); err != nil {
				return err
			}
		}
		node.client = csi.NewNodeClient(ClientConn)
		return nil
	},
}

// debug: go run . node get-info --endpoint 127.0.0.1:10000
var nodeGetInfoCmd = &cobra.Command{
	Use:     "get-info",
	Aliases: []string{"info"},
	Short:   `invokes the rpc "NodeGetInfo"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		rep, err := node.client.NodeGetInfo(context.TODO(), &csi.NodeGetInfoRequest{})
		if err != nil {
			return err
		}

		klog.Infof("NodeGetInfo response: %v", rep)

		return nil
	},
}

var nodePublishVolume struct {
	targetPath        string
	stagingTargetPath string
	pubCtx            mapOfStringArg
	volCtx            mapOfStringArg
	attribs           mapOfStringArg
	readOnly          bool
	caps              volumeCapabilitySliceArg
}

// debug: go run . node publish --target-path /mnt/hostpath1 --cap SINGLE_NODE_WRITER,block 335cf2b4-9edd-46fa-bfd5-af1db124ddf1 --endpoint 127.0.0.1:10000
var nodePublishVolumeCmd = &cobra.Command{
	Use:     "publish",
	Aliases: []string{"mnt", "mount"},
	Short:   `invokes the rpc "NodePublishVolume"`,
	Example: `
USAGE
    csc node publish [flags] VOLUME_ID [VOLUME_ID...]
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		req := csi.NodePublishVolumeRequest{
			StagingTargetPath: nodePublishVolume.stagingTargetPath,
			TargetPath:        nodePublishVolume.targetPath,
			PublishContext:    nodePublishVolume.pubCtx.data,
			Readonly:          nodePublishVolume.readOnly,
			Secrets: map[string]string{
				"secret": "abc123",
			},
			VolumeContext: nodePublishVolume.volCtx.data,
		}

		if len(nodePublishVolume.caps.data) > 0 {
			req.VolumeCapability = nodePublishVolume.caps.data[0]
		}

		for i := range args {
			// Set the volume ID for the current request.
			req.VolumeId = args[i]

			klog.Infof("mounting volume %v", req)
			_, err := node.client.NodePublishVolume(context.TODO(), &req)
			if err != nil {
				return err
			}

			klog.Infof("mounted volume %s", args[i])
		}

		return nil
	},
}

var nodeUnpublishVolume struct {
	targetPath string
}

// debug: go run . node unpublish --target-path /mnt/hostpath1 335cf2b4-9edd-46fa-bfd5-af1db124ddf1 --endpoint 127.0.0.1:10000
var nodeUnpublishVolumeCmd = &cobra.Command{
	Use:     "unpublish",
	Aliases: []string{"umount", "unmount"},
	Short:   `invokes the rpc "NodeUnpublishVolume"`,
	Example: `
USAGE
    csc node unpublish [flags] VOLUME_ID [VOLUME_ID...]
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		req := csi.NodeUnpublishVolumeRequest{
			TargetPath: nodeUnpublishVolume.targetPath,
		}

		for i := range args {
			// Set the volume ID for the current request.
			req.VolumeId = args[i]

			klog.Infof("unmounting volume %v", req)
			_, err := node.client.NodeUnpublishVolume(context.TODO(), &req)
			if err != nil {
				return err
			}

			klog.Infof("unmounted volume %s", args[i])
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(nodeCmd)

	nodeCmd.AddCommand(nodeGetInfoCmd)

	nodeCmd.AddCommand(nodePublishVolumeCmd)
	nodePublishVolumeCmd.Flags().StringVar(&nodePublishVolume.stagingTargetPath, "staging-target-path", "",
		"The path to which to stage or unstage the volume")
	nodePublishVolumeCmd.Flags().StringVar(&nodePublishVolume.targetPath, "target-path", "",
		"The path to which to mount or unmount the volume")
	nodePublishVolumeCmd.Flags().BoolVar(&nodePublishVolume.readOnly, "read-only", false, "Mark the volume as read-only")
	nodePublishVolumeCmd.Flags().Var(&nodePublishVolume.caps, "cap", `One or more volume capabilities may be specified using the following
        format:
			ACCESS_MODE,ACCESS_TYPE[,FS_TYPE,MOUNT_FLAGS]
        The ACCESS_MODE and ACCESS_TYPE values are required. Their values
        may be the their string name or their gRPC integer value. For example,
        the following two options are equivalent:
            --cap 5,1
            --cap MULTI_NODE_MULTI_WRITER,block
        If the access type specified is "mount" (or its gRPC field value of 2)
        then it's possible to specify a filesystem type and mount flags for
        the volume capability. Multiple mount flags may be specified using
        commas. For example:
            --cap MULTI_NODE_MULTI_WRITER,mount,xfs,uid=500,gid=500`)

	nodeCmd.AddCommand(nodeUnpublishVolumeCmd)
	nodeUnpublishVolumeCmd.Flags().StringVar(&nodeUnpublishVolume.targetPath, "target-path", "",
		"The path to which to mount or unmount the volume")
}
