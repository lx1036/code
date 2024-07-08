package endpoint

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"time"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/controller"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/endpoint/regeneration"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/eventqueue"
)

// startRegenerationFailureHandler waits for a build of the Endpoint to fail.
// Terminates when the given Endpoint is deleted.
// If a build fails, the controller tries to regenerate the
// Endpoint until it succeeds. Once the controller succeeds, it will not be
// ran again unless another build failure occurs. If the call to `Regenerate`
// fails inside of the controller,
func (e *Endpoint) startRegenerationFailureHandler() {
	e.controllers.UpdateController(fmt.Sprintf("endpoint-%s-regeneration-recovery", e.StringID()), controller.ControllerParams{
		DoFunc: func(ctx context.Context) error {
			select {
			case <-e.regenFailedChan:
				e.getLogger().Debug("received signal that regeneration failed")
			case <-ctx.Done():
				e.getLogger().Debug("exiting retrying regeneration goroutine due to endpoint being deleted")
				return nil
			}

			regenMetadata := &regeneration.ExternalRegenerationMetadata{
				// TODO (ianvernon) - is there a way we can plumb a parent
				// context to a controller (e.g., endpoint.aliveCtx)?
				ParentContext: ctx,
				Reason:        reasonRegenRetry,
				// Completely rewrite the endpoint - we don't know the nature
				// of the failure, simply that something failed.
				RegenerationLevel: regeneration.RegenerateWithDatapathRewrite,
			}
			regen, _ := e.SetRegenerateStateIfAlive(regenMetadata)
			if !regen {
				// We don't need to regenerate because the endpoint is d
				// disconnecting / is disconnected, or another regeneration has
				// already been enqueued. Exit gracefully.
				return nil
			}

			if success := <-e.Regenerate(regenMetadata); success {
				return nil
			}
			return fmt.Errorf("regeneration recovery failed")
		},
		ErrorRetryBaseDuration: 2 * time.Second,
		Context:                e.aliveCtx,
	})
}

// Regenerate forces the regeneration of endpoint programs & policy
// Should only be called with e.state at StateWaitingToRegenerate,
// StateWaitingForIdentity, or StateRestoring
func (e *Endpoint) Regenerate(regenMetadata *regeneration.ExternalRegenerationMetadata) <-chan bool {
	done := make(chan bool, 1)

	var (
		ctx   context.Context
		cFunc context.CancelFunc
	)

	if regenMetadata.ParentContext != nil {
		ctx, cFunc = context.WithCancel(regenMetadata.ParentContext)
	} else {
		ctx, cFunc = context.WithCancel(e.aliveCtx)
	}

	regenContext := ParseExternalRegenerationMetadata(ctx, cFunc, regenMetadata)

	epEvent := eventqueue.NewEvent(&EndpointRegenerationEvent{
		regenContext: regenContext,
		ep:           e,
	})

	// This may block if the Endpoint's EventQueue is full. This has to be done
	// synchronously as some callers depend on the fact that the event is
	// synchronously enqueued.
	resChan, err := e.eventQueue.Enqueue(epEvent)
	if err != nil {
		e.getLogger().Errorf("enqueue of EndpointRegenerationEvent failed: %s", err)
		done <- false
		close(done)
		return done
	}

	go func() {

		// Free up resources with context.
		defer cFunc()

		var (
			buildSuccess bool
			regenError   error
			canceled     bool
		)

		select {
		case result, ok := <-resChan:
			if ok {
				regenResult := result.(*EndpointRegenerationResult)
				regenError = regenResult.err
				buildSuccess = regenError == nil

				if regenError != nil {
					e.getLogger().WithError(regenError).Error("endpoint regeneration failed")
				}
			} else {
				// This may be unnecessary(?) since 'closing' of the results
				// channel means that event has been cancelled?
				e.getLogger().Debug("regeneration was cancelled")
				canceled = true
			}
		}

		// If a build is canceled, that means that the Endpoint is being deleted
		// not that the build failed.
		if !buildSuccess && !canceled {
			select {
			case e.regenFailedChan <- struct{}{}:
			default:
				// If we can't write to the channel, that means that it is
				// full / a regeneration will occur - we don't have to
				// do anything.
			}
		}
		done <- buildSuccess
		close(done)
	}()

	return done
}

// INFO: 这里会给当前 endpoint 做 regenerate bpf
func (e *Endpoint) regenerate(context *regenerationContext) (retErr error) {

	revision, stateDirComplete, err = e.regenerateBPF(context)
	if err != nil {
		failDir := e.FailedDirectoryPath()
		e.getLogger().WithFields(logrus.Fields{
			logfields.Path: failDir,
		}).Warn("generating BPF for endpoint failed, keeping stale directory.")

		// Remove an eventual existing previous failure directory
		e.removeDirectory(failDir)
		os.Rename(tmpDir, failDir)
		return err
	}

}
