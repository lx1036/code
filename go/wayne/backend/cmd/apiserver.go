package cmd

import "github.com/spf13/cobra"

var (
    APIServerCmd = &cobra.Command{
        Use:    "apiserver",
        PreRun: preRun,
        Run:    runApiServer,
    }
)

func preRun(cmd *cobra.Command, args []string) {
}


func runApiServer(cmd *cobra.Command, args []string) {

}
