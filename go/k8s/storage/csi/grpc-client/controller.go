package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

var controller struct {
	client csi.ControllerClient
}

// volumeCapabilitySliceArg is used for parsing one or more volume
// capabilities from the command line
type volumeCapabilitySliceArg struct {
	data []*csi.VolumeCapability
}

func (s *volumeCapabilitySliceArg) String() string {
	return ""
}

func (s *volumeCapabilitySliceArg) Type() string {
	return "mode,type[,fstype,mntflags]"
}

func (s *volumeCapabilitySliceArg) Set(val string) error {
	// The data can be split into a max of 4 components:
	// 1. mode
	// 2. cap
	// 3. fsType (if cap is mount)
	// 4. mntFlags (if cap is mount)
	data := strings.SplitN(val, ",", 4)
	if len(data) < 2 {
		return fmt.Errorf("invalid volume capability: %s", val)
	}

	var cap csi.VolumeCapability

	szMode := data[0]
	if i, ok := csi.VolumeCapability_AccessMode_Mode_value[szMode]; ok {
		cap.AccessMode = &csi.VolumeCapability_AccessMode{
			Mode: csi.VolumeCapability_AccessMode_Mode(i),
		}
	} else {
		i, err := strconv.ParseInt(szMode, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid access mode: %v: %v", szMode, err)
		}
		if _, ok := csi.VolumeCapability_AccessMode_Mode_name[int32(i)]; ok {
			cap.AccessMode = &csi.VolumeCapability_AccessMode{
				Mode: csi.VolumeCapability_AccessMode_Mode(i),
			}
		}
	}

	szType := data[1]

	// Handle a block volume capability
	if szType == "1" || strings.EqualFold(szType, "block") {
		cap.AccessType = &csi.VolumeCapability_Block{
			Block: &csi.VolumeCapability_BlockVolume{},
		}
		s.data = append(s.data, &cap)
		return nil
	}

	// Handle a mount volume capability
	if szType == "2" || strings.EqualFold(szType, "mount") {
		if len(data) < 3 {
			return fmt.Errorf("invalid volume capability: %s", val)
		}
		mountCap := &csi.VolumeCapability_MountVolume{
			FsType: data[2],
		}
		cap.AccessType = &csi.VolumeCapability_Mount{
			Mount: mountCap,
		}

		// If there is data remaining then treat it as mount flags.
		if len(data) > 3 {
			mountCap.MountFlags = strings.Split(data[3], ",")
		}

		s.data = append(s.data, &cap)
		return nil
	}

	return fmt.Errorf("invalid volume capability: %s", val)
}

// mapOfStringArg is used for parsing a csv, key=value arg into
// a map[string]string
type mapOfStringArg struct {
	sync.Once
	data map[string]string
}

func (s *mapOfStringArg) String() string {
	return ""
}

func (s *mapOfStringArg) Type() string {
	return "key=val[,key=val,...]"
}

func (s *mapOfStringArg) Set(val string) error {
	s.Do(func() { s.data = map[string]string{} })
	data := strings.Split(val, ",")
	for _, v := range data {
		vp := strings.SplitN(v, "=", 2)
		switch len(vp) {
		case 1:
			s.data[vp[0]] = ""
		case 2:
			s.data[vp[0]] = vp[1]
		}
	}
	return nil
}

var createVolume struct {
	reqBytes int64
	limBytes int64
	caps     volumeCapabilitySliceArg
	params   mapOfStringArg
}

// controllerCmd represents the controller command
var controllerCmd = &cobra.Command{
	Use:     "controller",
	Aliases: []string{"c"},
	Short:   "the csi controller service rpcs",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if f := cmd.Root().PersistentPreRunE; f != nil {
			if err := f(cmd, args); err != nil {
				return err
			}
		}
		controller.client = csi.NewControllerClient(ClientConn)
		return nil
	},
}

// debug: go run . controller create-volume --cap 1,mount,xfs,uid=500 myvolume1 myvolume2 --endpoint 127.0.0.1:10000
var createVolumeCmd = &cobra.Command{
	Use:     "create-volume",
	Aliases: []string{"new"},
	Example: `
		CREATING MULTIPLE VOLUMES
        The following example illustrates how to create two volumes with the
        same characteristics at the same time:
            csc controller new --endpoint /csi/server.sock
                               --cap 1,block \
                               --cap MULTI_NODE_MULTI_WRITER,mount,xfs,uid=500 \
                               --params region=us,zone=texas
                               --params disabled=false
                               MyNewVolume1 MyNewVolume2
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		req := csi.CreateVolumeRequest{
			VolumeCapabilities: createVolume.caps.data,
			Parameters:         createVolume.params.data,
			Secrets: map[string]string{
				"secret": "abc123",
			},
		}
		if createVolume.reqBytes > 0 {
			req.CapacityRange.RequiredBytes = createVolume.reqBytes
		}
		if createVolume.limBytes > 0 {
			req.CapacityRange.LimitBytes = createVolume.limBytes
		}

		klog.Infof("create %d number volumes", len(args))

		for i := range args {
			// Set the volume name for the current request.
			req.Name = args[i]

			klog.Infof("creating volume request %v", req)
			response, err := controller.client.CreateVolume(context.TODO(), &req)
			if err != nil {
				return err
			}

			klog.Infof("CreateVolume response: %v", response)
		}

		return nil
	},
}

// debug: go run . controller delete-volume a53dd461-634f-4dbb-a10c-38de39de4396 --endpoint 127.0.0.1:10000
var deleteVolumeCmd = &cobra.Command{
	Use:     "delete-volume",
	Aliases: []string{"d", "rm", "del", "delete"},
	Short:   `invokes the rpc "DeleteVolume"`,
	Example: `
USAGE
    csc controller deletevolume [flags] VOLUME_ID [VOLUME_ID...]
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		req := csi.DeleteVolumeRequest{
			Secrets: map[string]string{
				"secret": "abc123",
			},
		}

		for i := range args {
			// Set the volume ID for the current request.
			req.VolumeId = args[i]

			klog.Infof("deleting volume request %v", req)
			_, err := controller.client.DeleteVolume(context.TODO(), &req)
			if err != nil {
				return err
			}

			klog.Infof("deleted volume id: %s", args[i])
		}

		return nil
	},
}

var createSnapshot struct {
	sourceVol string
	params    mapOfStringArg
}

// debug: go run . controller create-snapshot --source-volume 335cf2b4-9edd-46fa-bfd5-af1db124ddf1 mysnap1 mysnap2 --endpoint 127.0.0.1:10000
var createSnapshotCmd = &cobra.Command{
	Use: "create-snapshot",
	Example: `
CREATING MULTIPLE SNAPSHOTS
        The following example illustrates how to create two snapshots with the
        same characteristics at the same time:
            csc controller snap --endpoint /csi/server.sock
							    --source-volume MySourceVolume
                                MyNewSnapshot1 MyNewSnapshot2
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		req := csi.CreateSnapshotRequest{
			SourceVolumeId: createSnapshot.sourceVol,
			Parameters:     createSnapshot.params.data,
			Secrets: map[string]string{
				"secret": "abc123",
			},
		}

		for i := range args {
			// Set the volume name for the current request.
			req.Name = args[i]
			if createSnapshot.sourceVol == "" {
				return fmt.Errorf("--source-volume MUST be provided")
			}

			klog.Infof("creating snapshot request %v", req)
			response, err := controller.client.CreateSnapshot(context.TODO(), &req)
			if err != nil {
				return err
			}

			klog.Infof("CreateSnapshot response: %v", response)

			return nil
		}

		return nil

	},
}

// debug: go run . controller delete-snapshot a58191fd-9791-4fc6-8e83-206dbb92a832 --endpoint 127.0.0.1:10000
var deleteSnapshotCmd = &cobra.Command{
	Use:     "delete-snapshot",
	Aliases: []string{"ds", "delsnap"},
	Short:   `invokes the rpc "DeleteSnapshot"`,
	Example: `
USAGE
    csc controller delete-snapshot [flags] snapshot_ID [snapshot_ID...]
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		req := csi.DeleteSnapshotRequest{
			Secrets: map[string]string{
				"secret": "abc123",
			},
		}

		for i := range args {
			// Set the snapshot ID for the current request.
			req.SnapshotId = args[i]

			klog.Infof("deleting snapshot request %v", req)
			_, err := controller.client.DeleteSnapshot(context.TODO(), &req)
			if err != nil {
				return err
			}

			klog.Infof("deleted snapshot id: %s", args[i])
		}

		return nil
	},
}

var valVolCaps struct {
	volCtx mapOfStringArg
	params mapOfStringArg
	caps   volumeCapabilitySliceArg
}

// debug: go run . controller validate-volume-capabilities --cap 1,mount,xfs,uid=500 335cf2b4-9edd-46fa-bfd5-af1db124ddf1 --endpoint 127.0.0.1:10000
var valVolCapsCmd = &cobra.Command{
	Use:     "validate-volume-capabilities",
	Aliases: []string{"validate"},
	Short:   `invokes the rpc "ValidateVolumeCapabilities"`,
	Example: `
USAGE
    csc controller validate [flags] VOLUME_ID [VOLUME_ID...]
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		req := csi.ValidateVolumeCapabilitiesRequest{
			VolumeContext:      valVolCaps.volCtx.data,
			VolumeCapabilities: valVolCaps.caps.data,
			Parameters:         valVolCaps.params.data,
		}

		for i := range args {
			// Set the volume name for the current request.
			req.VolumeId = args[i]

			klog.Infof("validating volume capabilities %v", req)
			rep, err := controller.client.ValidateVolumeCapabilities(context.TODO(), &req)
			if err != nil {
				return err
			}

			klog.Infof("validated volume capabilities %s with confirmed: %v", rep.Message, rep.Confirmed)
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(controllerCmd)

	controllerCmd.AddCommand(createVolumeCmd)
	createVolumeCmd.Flags().Int64Var(&createVolume.reqBytes, "req-bytes", 0, "The required size of the volume in bytes")
	createVolumeCmd.Flags().Int64Var(&createVolume.limBytes, "lim-bytes", 0, "The limit to the size of the volume in bytes")
	createVolumeCmd.Flags().Var(&createVolume.params, "params", `One or more key/value pairs may be specified to send with
        the request as its Parameters field:
            --params key1=val1,key2=val2 --params=key3=val3`)
	createVolumeCmd.Flags().Var(&createVolume.caps, "cap", `One or more volume capabilities may be specified using the following
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

	controllerCmd.AddCommand(createSnapshotCmd)
	createSnapshotCmd.Flags().StringVar(&createSnapshot.sourceVol, "source-volume", "", "The source volume to snapshot")
	createSnapshotCmd.Flags().Var(&createSnapshot.params, "params", `
		One or more key/value pairs may be specified to send with
        the request as its Parameters field:
            --params key1=val1,key2=val2 --params=key3=val3
	`)

	controllerCmd.AddCommand(deleteVolumeCmd)
	controllerCmd.AddCommand(deleteSnapshotCmd)

	controllerCmd.AddCommand(valVolCapsCmd)
	valVolCapsCmd.Flags().Var(&valVolCaps.caps, "cap", `One or more volume capabilities may be specified using the following
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
	valVolCapsCmd.Flags().Var(&valVolCaps.volCtx, "vol-context", `One or more key/value pairs may be specified to send with
        the request as its VolumeContext field:
            --vol-context key1=val1,key2=val2 --vol-context=key3=val3`)
	valVolCapsCmd.Flags().Var(&valVolCaps.params, "params", `
		One or more key/value pairs may be specified to send with
        the request as its Parameters field:
            --params key1=val1,key2=val2 --params=key3=val3
	`)
}
