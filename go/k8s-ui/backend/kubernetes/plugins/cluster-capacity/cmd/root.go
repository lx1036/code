package cmd

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "cd",
		Short: "cd is cmd for continuous deployment",
		Long:  `cd`,
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(clusterCapacityCmd)
}
