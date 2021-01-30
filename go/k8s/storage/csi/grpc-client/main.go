package main

import (
	"flag"
	"fmt"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"html/template"
)

var (
	ClientConn *grpc.ClientConn
	Tpl        *template.Template

	endpoint = flag.String("endpoint", "localhost:10000", "")

	RootCmd = &cobra.Command{
		Use: "grpc-client",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			clientConn, err := grpc.Dial(*endpoint, grpc.WithInsecure())
			if err != nil {
				return err
			}

			ClientConn = clientConn

			var format string
			switch cmd.Name() {
			case pluginInfoCmd.Name():
				format = pluginInfoFormat
			}
			tpl, err := template.New("t").Funcs(template.FuncMap{
				"isa": func(o interface{}, t string) bool {
					return fmt.Sprintf("%T", o) == t
				},
			}).Parse(format)
			if err != nil {
				return err
			}

			Tpl = tpl

			return nil
		},
	}
)

// debug: go run . identity plugin-info
func main() {
	flag.Parse()

	flag.Set("logtostderr", "true")

	err := RootCmd.Execute()
	if err != nil {
		glog.Error(err)
	}
}
