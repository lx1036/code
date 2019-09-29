package cmd

import (
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/spf13/cobra"
	"k8s-lx1036/wayne/backend/initial"
	"sync"
)

var Version string

var RootCmd = &cobra.Command{
	Use: "wayne",
}

var VersionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("wayne %s \n", Version)
	},
}

var APIServerCmd = &cobra.Command{
		Use:    "apiserver",
		PreRun: preRun,
		Run:    runApiServer,
	}

func preRun(cmd *cobra.Command, args []string) {
}

func runApiServer(cmd *cobra.Command, args []string) {
	// MySQL
	initial.InitDb()

	// Swagger API
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	// K8S Client
	initial.InitClient()

	// 初始化RabbitMQ
	busEnable := beego.AppConfig.DefaultBool("BusEnable", false)
	if busEnable {
		initial.InitBus()
	}

	// 初始化RsaPrivateKey
	initial.InitRsaKey()

	// init kube labels
	initial.InitKubeLabel()

	beego.Run()
}

var (
	WorkerCmd = &cobra.Command{
		Use:     "worker",
		PreRunE: preRunE,
		Run:     runWorker,
	}

	workerType        string
	concurrency       int
	lock              sync.Mutex
	availableRecovery = 3
)

func preRunE(cmd *cobra.Command, args []string) error {
	if workerType == "" {
		return errors.New("missing worker type")
	}

	switch workerType {
	case "AuditWorker", "WebhookWorker":
		break
	default:
		return errors.New(fmt.Sprintf("unknown worker type: %s", workerType))
	}

	return nil
}

func runWorker(cmd *cobra.Command, args []string) {

}

func init() {
	cobra.EnableCommandSorting = false

	RootCmd.AddCommand(APIServerCmd)
	RootCmd.AddCommand(WorkerCmd)
	RootCmd.AddCommand(VersionCmd)
}
