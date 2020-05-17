package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s-lx1036/k8s-ui/backend/kubernetes/docker/write-my-docker/container"
)

var (
	InitCmd = &cobra.Command{
		Use:   "init",
		Short: "Init container process run user's process in container. Do not call it outside",
		Run: func(cmd *cobra.Command, args []string) {
			log.Info("init come on")
			_ = container.RunContainerInitProcess()
		},
	}
)

func init() {

}
