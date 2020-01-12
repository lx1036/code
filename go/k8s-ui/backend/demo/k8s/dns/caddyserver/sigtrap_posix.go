package caddy

import (
	"github.com/mholt/certmagic"
	"k8s-lx1036/k8s-ui/backend/demo/k8s/dns/caddyserver/telemetry"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// trapSignalsPosix captures POSIX-only signals.
func trapSignalsPosix() {
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)

		for sig := range sigchan {
			switch sig {
			case syscall.SIGQUIT:
				log.Println("[INFO] SIGQUIT: Quitting process immediately")
				for _, f := range OnProcessExit {
					f() // only perform important cleanup actions
				}
				certmagic.CleanUpOwnLocks()
				os.Exit(0)

			case syscall.SIGTERM:
				log.Println("[INFO] SIGTERM: Shutting down servers then terminating")
				exitCode := executeShutdownCallbacks("SIGTERM")
				for _, f := range OnProcessExit {
					f() // only perform important cleanup actions
				}
				err := Stop()
				if err != nil {
					log.Printf("[ERROR] SIGTERM stop: %v", err)
					exitCode = 3
				}

				telemetry.AppendUnique("sigtrap", "SIGTERM")
				go telemetry.StopEmitting() // won't finish in time, but that's OK - just don't block

				certmagic.CleanUpOwnLocks()
				os.Exit(exitCode)

			case syscall.SIGUSR1:
				log.Println("[INFO] SIGUSR1: Reloading")
				go telemetry.AppendUnique("sigtrap", "SIGUSR1")

				// Start with the existing Caddyfile
				caddyfileToUse, inst, err := getCurrentCaddyfile()
				if err != nil {
					log.Printf("[ERROR] SIGUSR1: %v", err)
					continue
				}
				if loaderUsed.loader == nil {
					// This also should never happen
					log.Println("[ERROR] SIGUSR1: no Caddyfile loader with which to reload Caddyfile")
					continue
				}

				// Load the updated Caddyfile
				newCaddyfile, err := loaderUsed.loader.Load(inst.serverType)
				if err != nil {
					log.Printf("[ERROR] SIGUSR1: loading updated Caddyfile: %v", err)
					continue
				}
				if newCaddyfile != nil {
					caddyfileToUse = newCaddyfile
				}

				// Backup old event hooks
				oldEventHooks := cloneEventHooks()

				// Purge the old event hooks
				purgeEventHooks()

				// Kick off the restart; our work is done
				EmitEvent(InstanceRestartEvent, nil)
				_, err = inst.Restart(caddyfileToUse)
				if err != nil {
					restoreEventHooks(oldEventHooks)

					log.Printf("[ERROR] SIGUSR1: %v", err)
				}

			case syscall.SIGUSR2:
				log.Println("[INFO] SIGUSR2: Upgrading")
				go telemetry.AppendUnique("sigtrap", "SIGUSR2")
				if err := Upgrade(); err != nil {
					log.Printf("[ERROR] SIGUSR2: upgrading: %v", err)
				}

			case syscall.SIGHUP:
				// ignore; this signal is sometimes sent outside of the user's control
				go telemetry.AppendUnique("sigtrap", "SIGHUP")
			}
		}
	}()
}
