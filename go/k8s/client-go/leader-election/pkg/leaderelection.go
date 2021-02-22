package pkg

import (
	"bytes"
	"context"
	"time"

	"k8s-lx1036/k8s/client-go/leader-election/pkg/resourcelock"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/clock"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

const (
	JitterFactor = 1.2
)

type LeaderCallbacks struct {
	// OnStartedLeading is called when a LeaderElector client starts leading
	OnStartedLeading func(context.Context)
	// OnStoppedLeading is called when a LeaderElector client stops leading
	OnStoppedLeading func()
	// OnNewLeader is called when the client observes a leader that is
	// not the previously observed leader. This includes the first observed
	// leader when the client starts.
	OnNewLeader func(identity string)
}

type LeaderElectionConfig struct {
	Lock resourcelock.Interface

	// Callbacks are callbacks that are triggered during certain lifecycle
	// events of the LeaderElector
	Callbacks LeaderCallbacks

	RetryPeriod time.Duration

	// LeaseDuration is the duration that non-leader candidates will
	// wait to force acquire leadership. This is measured against time of
	// last observed ack.
	LeaseDuration time.Duration

	RenewDeadline time.Duration

	ReleaseOnCancel bool
}

type LeaderElector struct {
	config LeaderElectionConfig

	// clock is wrapper around time to allow for less flaky testing
	clock clock.Clock

	observedRawRecord []byte
	observedRecord    resourcelock.LeaderElectionRecord
	observedTime      time.Time

	reportedLeader string
}

// 通过判断HolderIdentity判定leader归属
func (leaderElector *LeaderElector) IsLeader() bool {
	return leaderElector.observedRecord.HolderIdentity == leaderElector.config.Lock.Identity()
}

// tryAcquireOrRenew tries to acquire a leader lease if it is not already acquired,
// else it tries to renew the lease if it has already been acquired.
func (leaderElector *LeaderElector) tryAcquireOrRenew(ctx context.Context) bool {
	now := metav1.Now()
	leaderElectionRecord := resourcelock.LeaderElectionRecord{
		HolderIdentity:       leaderElector.config.Lock.Identity(),
		LeaseDurationSeconds: int(leaderElector.config.LeaseDuration / time.Second),
		AcquireTime:          now,
		RenewTime:            now,
	}

	// 1. obtain or create the ElectionRecord
	oldLeaderElectionRecord, oldLeaderElectionRawRecord, err := leaderElector.config.Lock.Get(ctx)
	if err != nil {
		if !errors.IsNotFound(err) {
			klog.Errorf("error retrieving resource lock %v: %v", leaderElector.config.Lock.Describe(), err)
			return false
		}

		if err = leaderElector.config.Lock.Create(ctx, leaderElectionRecord); err != nil {
			klog.Errorf("error initially creating leader election record: %v", err)
			return false
		}

		leaderElector.observedRecord = leaderElectionRecord
		leaderElector.observedTime = leaderElector.clock.Now()
		return true
	}

	// 2. Record obtained, check the Identity & Time
	if !bytes.Equal(leaderElector.observedRawRecord, oldLeaderElectionRawRecord) {
		leaderElector.observedRecord = *oldLeaderElectionRecord
		leaderElector.observedRawRecord = oldLeaderElectionRawRecord
		leaderElector.observedTime = leaderElector.clock.Now()
	}

	// 3. We're going to try to update. The leaderElectionRecord is set to it's default
	// here. Let's correct it before updating.
	if leaderElector.IsLeader() {
		leaderElectionRecord.AcquireTime = oldLeaderElectionRecord.AcquireTime
		leaderElectionRecord.LeaderTransitions = oldLeaderElectionRecord.LeaderTransitions
	} else {
		leaderElectionRecord.LeaderTransitions = oldLeaderElectionRecord.LeaderTransitions + 1
	}

	// update the lock itself
	if err = leaderElector.config.Lock.Update(ctx, leaderElectionRecord); err != nil {
		klog.Errorf("Failed to update lock: %v", err)
		return false
	}

	leaderElector.observedRecord = leaderElectionRecord
	leaderElector.observedTime = leaderElector.clock.Now()

	return true
}

// 执行OnNewLeader callback
func (leaderElector *LeaderElector) maybeReportTransition() {
	if leaderElector.observedRecord.HolderIdentity == leaderElector.reportedLeader {
		return
	}

	leaderElector.reportedLeader = leaderElector.observedRecord.HolderIdentity
	if leaderElector.config.Callbacks.OnNewLeader != nil {
		go leaderElector.config.Callbacks.OnNewLeader(leaderElector.reportedLeader)
	}
}

func (leaderElector *LeaderElector) acquire(ctx context.Context) bool {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	succeeded := false
	desc := leaderElector.config.Lock.Describe()

	wait.JitterUntil(func() {
		succeeded = leaderElector.tryAcquireOrRenew(ctx)
		leaderElector.maybeReportTransition()
		if !succeeded {
			klog.V(4).Infof("failed to acquire lease %v", desc)
			return
		}

		klog.Infof("successfully acquired lease %v", desc)
		// 如果succeeded了，则关闭定时任务
		cancel()
	}, leaderElector.config.RetryPeriod, JitterFactor, true, ctx.Done())

	return succeeded
}

func (leaderElector *LeaderElector) renew(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	desc := leaderElector.config.Lock.Describe()
	wait.Until(func() {
		timeoutCtx, timeoutCancel := context.WithTimeout(ctx, leaderElector.config.RenewDeadline)
		defer timeoutCancel()

		err := wait.PollImmediateUntil(leaderElector.config.RetryPeriod, func() (done bool, err error) {
			return leaderElector.tryAcquireOrRenew(timeoutCtx), nil
		}, timeoutCtx.Done())
		leaderElector.maybeReportTransition()
		if err == nil {
			klog.V(5).Infof("successfully renewed lease %v", desc)
			return
		}

		klog.Infof("failed to renew lease %v: %v", desc, err)
		cancel()
	}, leaderElector.config.RetryPeriod, ctx.Done())

	// if we hold the lease, give it up
	if leaderElector.config.ReleaseOnCancel {
		//leaderElector.release()
	}
}

func (leaderElector *LeaderElector) Run(ctx context.Context) {
	defer func() {
		runtime.HandleCrash()
		leaderElector.config.Callbacks.OnStoppedLeading()
	}()

	// 会阻塞在这个逻辑，定时任务会一直尝试获取leader
	if !leaderElector.acquire(ctx) {
		return // ctx signalled done
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 执行自定义逻辑
	go leaderElector.config.Callbacks.OnStartedLeading(ctx)

	leaderElector.renew(ctx)
}

func RunOrDie(ctx context.Context, config LeaderElectionConfig) {
	leaderElector := &LeaderElector{
		config: config,
		clock:  clock.RealClock{},
	}

	leaderElector.Run(ctx)
}
