package main

import (
	"flag"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"
)

// @see https://github.com/rexray/gocsi/blob/master/csc/README.md

var (
	ClientConn *grpc.ClientConn

	endpoint string

	//endpoint = flag.String("endpoint", "localhost:10000", "")

	RootCmd = &cobra.Command{
		Use: "grpc-client",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			clientConn, err := grpc.Dial(endpoint, grpc.WithInsecure())
			if err != nil {
				return err
			}

			ClientConn = clientConn

			/*var format string
			switch cmd.Name() {
			case pluginInfoCmd.Name():
				format = pluginInfoFormat
			case pluginCapabilityCmd.Name():
				format = pluginCapsFormat
			}
			tpl, err := template.New("t").Funcs(template.FuncMap{
				"isa": func(o interface{}, t string) bool {
					return fmt.Sprintf("%T", o) == t
				},
			}).Parse(format)
			if err != nil {
				return err
			}

			Tpl = tpl*/

			return nil
		},
	}
)

func init() {
	RootCmd.PersistentFlags().StringVarP(&endpoint, "endpoint", "e", "127.0.0.1:10000", `The CSI endpoint may also be specified by the environment variable
        CSI_ENDPOINT. The endpoint should adhere to Go's network address
        pattern:
            * tcp://host:port
            * unix:///path/to/file.sock.
        If the network type is omitted then the value is assumed to be an
        absolute or relative filesystem path to a UNIX socket file`)
}

// go run . xxx --endpoint 127.0.0.1:10000
func main() {
	flag.Set("logtostderr", "true")
	flag.Parse()

	err := RootCmd.Execute()
	if err != nil {
		klog.Error(err)
	}
}
