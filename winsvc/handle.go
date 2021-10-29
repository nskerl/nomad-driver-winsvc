package winsvc

import (
	"context"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/client/stats"
	shelpers "github.com/hashicorp/nomad/helper/stats"
	"github.com/hashicorp/nomad/plugins/drivers"
)

type taskHandle struct {
	logger hclog.Logger

	// stateLock syncs access to all fields below
	stateLock sync.RWMutex
	//procState   drivers.TaskState
	taskConfig  *drivers.TaskConfig
	state       drivers.TaskState
	serviceName string
	startedAt   time.Time
	completedAt time.Time
	exitResult  *drivers.ExitResult

	client        ServiceClientInterface
	cpuStatsSys   *stats.CpuStats
	cpuStatsUser  *stats.CpuStats
	cpuStatsTotal *stats.CpuStats
}

func (h *taskHandle) TaskStatus() *drivers.TaskStatus {
	h.stateLock.RLock()
	defer h.stateLock.RUnlock()

	return &drivers.TaskStatus{
		ID:               h.taskConfig.ID,
		Name:             h.taskConfig.Name,
		State:            h.state,
		StartedAt:        h.startedAt,
		CompletedAt:      h.completedAt,
		ExitResult:       h.exitResult,
		DriverAttributes: map[string]string{
			// no custom attributes needed
		},
	}
}

func (h *taskHandle) IsRunning() bool {
	h.stateLock.RLock()
	defer h.stateLock.RUnlock()
	return h.state == drivers.TaskStateRunning
}

func (h *taskHandle) run() {
	h.stateLock.Lock()
	if h.exitResult == nil {
		h.exitResult = &drivers.ExitResult{}
	}

	h.stateLock.Unlock()

	if err := shelpers.Init(); err != nil {
		h.logger.Error("unable to initialize stats", "error", err)
	}

	for {
		time.Sleep(5 * time.Second)

		isRunning, err := h.client.IsServiceRunning(h.serviceName)
		if err != nil {
			break
		}

		if !isRunning {
			break
		}
	}
	h.stateLock.Lock()
	defer h.stateLock.Unlock()

	h.state = drivers.TaskStateExited
	h.exitResult.ExitCode = 0
	h.exitResult.Signal = 0
	h.completedAt = time.Now()
}

func (h *taskHandle) stats(ctx context.Context, ch chan *drivers.TaskResourceUsage, interval time.Duration) {
	defer close(ch)
	timer := time.NewTimer(0)
	for {
		select {
		case <-ctx.Done():
			return

		case <-timer.C:
			timer.Reset(interval)
		}

		serviceStats, err := h.client.GetServiceStats(h.serviceName)
		if err != nil || serviceStats == nil {
			h.logger.Warn("Did not get service process stats:", "service", h.serviceName, "error", err)
			return
		}

		select {
		case <-ctx.Done():
			return
		case ch <- h.getTaskResourceUsage(serviceStats):
		}
	}
}

func (h *taskHandle) getTaskResourceUsage(stats *wmiProcessStats) *drivers.TaskResourceUsage {

	totalPercent := h.cpuStatsTotal.Percent(float64(stats.KernelModeTime + stats.UserModeTime))

	cs := &drivers.CpuStats{
		SystemMode: h.cpuStatsSys.Percent(float64(stats.KernelModeTime)),
		UserMode:   h.cpuStatsUser.Percent(float64(stats.UserModeTime)),
		Percent:    totalPercent,
		Measured:   []string{"Percent", "System Mode", "User Mode"},
		TotalTicks: h.cpuStatsTotal.TicksConsumed(totalPercent),
	}

	ms := &drivers.MemoryStats{
		RSS:      stats.WorkingSetPrivate,
		Measured: []string{"RSS"},
	}

	ts := time.Now().UTC().UnixNano()

	return &drivers.TaskResourceUsage{
		ResourceUsage: &drivers.ResourceUsage{
			CpuStats:    cs,
			MemoryStats: ms,
		},
		Timestamp: ts,
	}
}

func (h *taskHandle) shutdown(timeout time.Duration) error {
	h.stateLock.Lock()
	defer h.stateLock.Unlock()

	h.logger.Debug("Calling stopService from shutdown", "service", h.serviceName, "timeout", timeout.String())

	if h.client == nil {
		h.logger.Warn("Client is nil")
	}

	// send a stop request to service
	if err := h.client.StopService(h.serviceName); err != nil {
		h.logger.Warn("Error while sending stop command", "service", h.serviceName, "error", err)
	}

	// wait for the service to stop gracefully
	time.Sleep(timeout)

	// stop any stubborn processes attached to service (if service already stopped cleanly this is a no-op)
	if err := h.client.KillService(h.serviceName); err != nil {
		h.logger.Warn("Error while killing tasks", "service", h.serviceName, "error", err)
	}

	return nil
}

func (h *taskHandle) cleanup() error {
	h.stateLock.Lock()

	defer h.stateLock.Unlock()

	if err := h.client.KillService(h.serviceName); err != nil {
		return err
	}

	if err := h.client.RemoveService(h.serviceName); err != nil {
		return err
	}

	return nil
}
