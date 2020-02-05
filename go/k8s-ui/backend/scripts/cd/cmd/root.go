package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
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

func er(msg interface{}) {
	fmt.Println("Error:", msg)
	os.Exit(1)
}

func init() {
	rootCmd.AddCommand(deployCmd)
}
