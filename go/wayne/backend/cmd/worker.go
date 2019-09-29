package cmd

import (
    "errors"
    "fmt"
    "github.com/spf13/cobra"
    "sync"
)

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
