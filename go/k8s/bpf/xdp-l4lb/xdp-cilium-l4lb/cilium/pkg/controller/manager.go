package controller

import (
	"context"
	"fmt"
	"time"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/lock"

	"github.com/google/uuid"
)

var (
	// globalStatus is the global status of all controllers
	globalStatus = NewManager()
)

type controllerMap map[string]*Controller

// Manager is a list of controllers
type Manager struct {
	controllers controllerMap
	mutex       lock.RWMutex
}

// NewManager allocates a new manager
func NewManager() *Manager {
	return &Manager{
		controllers: controllerMap{},
	}
}

// UpdateController installs or updates a controller in the manager. A
// controller is identified by its name. If a controller with the name already
// exists, the controller will be shut down and replaced with the provided
// controller. Updating a controller will cause the DoFunc to be run
// immediately regardless of any previous conditions. It will also cause any
// statistics to be reset.
func (m *Manager) UpdateController(name string, params ControllerParams) {
	m.updateController(name, params)
}

func (m *Manager) updateController(name string, params ControllerParams) *Controller {
	start := time.Now()

	// ensure the callbacks are valid
	if params.DoFunc == nil {
		params.DoFunc = func(ctx context.Context) error {
			return fmt.Errorf("controller %s DoFunc is nil", name)
		}
	}
	if params.StopFunc == nil {
		params.StopFunc = NoopFunc
	}

	m.mutex.Lock()

	if m.controllers == nil {
		m.controllers = controllerMap{}
	}

	ctrl, exists := m.controllers[name]
	if exists {
		m.mutex.Unlock()

		ctrl.getLogger().Debug("Updating existing controller")
		ctrl.mutex.Lock()
		ctrl.updateParamsLocked(params)
		ctrl.mutex.Unlock()

		// Notify the goroutine of the params update.
		select {
		case ctrl.update <- struct{}{}:
		default:
		}

		ctrl.getLogger().Debug("Controller update time: ", time.Since(start))
	} else {
		ctrl = &Controller{
			name:       name,
			uuid:       uuid.New().String(),
			stop:       make(chan struct{}),
			update:     make(chan struct{}, 1),
			trigger:    make(chan struct{}, 1),
			terminated: make(chan struct{}),
		}
		ctrl.updateParamsLocked(params)
		ctrl.getLogger().Debug("Starting new controller")

		if params.Context == nil {
			ctrl.ctxDoFunc, ctrl.cancelDoFunc = context.WithCancel(context.Background())
		} else {
			ctrl.ctxDoFunc, ctrl.cancelDoFunc = context.WithCancel(params.Context)
		}
		m.controllers[ctrl.name] = ctrl
		m.mutex.Unlock()

		globalStatus.mutex.Lock()
		globalStatus.controllers[ctrl.uuid] = ctrl
		globalStatus.mutex.Unlock()

		go ctrl.runController()
	}

	return ctrl
}

// RemoveController stops and removes a controller from the manager. If DoFunc
// is currently running, DoFunc is allowed to complete in the background.
func (m *Manager) RemoveController(name string) error {
	_, err := m.removeAndReturnController(name)
	return err
}

func (m *Manager) removeAndReturnController(name string) (*Controller, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.controllers == nil {
		return nil, fmt.Errorf("empty controller map")
	}

	oldCtrl, ok := m.controllers[name]
	if !ok {
		return nil, fmt.Errorf("unable to find controller %s", name)
	}

	m.removeController(oldCtrl)

	return oldCtrl, nil
}

func (m *Manager) removeController(ctrl *Controller) {
	ctrl.stopController()
	delete(m.controllers, ctrl.name)

	globalStatus.mutex.Lock()
	delete(globalStatus.controllers, ctrl.uuid)
	globalStatus.mutex.Unlock()

	ctrl.getLogger().Debug("Removed controller")
}
