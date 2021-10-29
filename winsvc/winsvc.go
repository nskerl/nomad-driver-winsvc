package winsvc

import (
	"fmt"
	"os/exec"
	"sync"
	"syscall"

	"github.com/StackExchange/wmi"
	"github.com/hashicorp/go-hclog"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

var mux sync.Mutex

type ServiceClient struct {
	logger hclog.Logger
}

type Win32Service struct {
	ExitCode  uint16
	Name      string
	ProcessId uint16
	StartMode string
	State     string
	Status    string
}

type ServiceClientInterface interface {
	CreateService(name string, executable string, username string, password string, args []string) error
	RemoveService(name string) error
	StartService(name string) error
	StopService(name string) error
	KillService(name string) error
	InspectService(name string) (*Win32Service, error)
	IsServiceRunning(name string) (bool, error)
	GetServiceStats(name string) (*wmiProcessStats, error)
}

// WMI Internal Type
type win32PerfFormattedDataPerfProcProcess struct {
	WorkingSetPrivate uint64
}

// WMI Internal Type
type win32Process struct {
	KernelModeTime uint64
	UserModeTime   uint64
}

type wmiProcessStats struct {
	KernelModeTime    uint64
	UserModeTime      uint64
	WorkingSetPrivate uint64
}

func (c ServiceClient) CreateService(name string, executable string, username string, password string, args []string) error {
	c.logger.Debug("Creating service", "name", name, "executable", executable)
	mux.Lock()
	defer mux.Unlock()

	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err == nil {
		s.Close()
		c.logger.Error("Service already exists", "name", name)
		return err
	}
	config := mgr.Config{DisplayName: name, Description: "Nomad managed service instance"}

	if username != "" && password != "" {
		config.ServiceStartName = username
		config.Password = password
	}

	s, err = m.CreateService(name, executable, config, args...)
	if err != nil {
		c.logger.Error("Error creating service", "name", name, "executable", executable, "error", err)
		return err
	}
	defer s.Close()

	return nil
}

func (c ServiceClient) StartService(name string) error {
	c.logger.Debug("Starting service", "name", name)
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		c.logger.Error("Error starting service", "name", name, "error", err)
		return err
	}
	defer s.Close()
	return s.Start()
}

func (c ServiceClient) StopService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		c.logger.Error("Could not connect to service control manager")
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		c.logger.Error("Failed to open service", "service", name, "error", err)
		return err
	}
	defer s.Close()

	c.logger.Debug("Sending stop command", "service", name)
	_, err = s.Control(svc.Stop)
	if err != nil {
		c.logger.Error("Failed to send stop command", "service", name, "error", err)
		return err
	}

	return nil
}

func (c ServiceClient) RemoveService(name string) error {
	c.logger.Debug("Remove service", "service", name)

	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		c.logger.Error("Failed to open service", "service", name, "error", err)
		return err
	}
	defer s.Close()

	err = s.Delete()
	if err != nil {
		c.logger.Error("Error while deleting service", "service", name, "error", err)
		return err
	}

	return nil
}

func (c ServiceClient) KillService(name string) error {

	output, err := exec.Command("taskkill", "/F", "/FI", fmt.Sprintf("SERVICES eq %s", name)).CombinedOutput()
	if err != nil {
		c.logger.Error("Failed to kill tasks for service", "service", name, "error", err)
		return err
	} else {
		c.logger.Debug("Killed tasks for service", "service", name, "output", string(output))
	}

	return nil
}

func (c ServiceClient) IsServiceRunning(name string) (bool, error) {

	m, err := mgr.Connect()
	if err != nil {
		return false, err
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		if errno, ok := err.(syscall.Errno); ok && errno == 1060 {
			c.logger.Error("Service does not exist", "service", name)
			return false, err
		}
		return false, err
	}
	defer s.Close()

	status, err := s.Query()
	if err != nil {
		c.logger.Error("Error running status query", "service", name)
		return false, err
	}

	switch status.State {
	case svc.StartPending:
		fallthrough
	case svc.Running:
		return true, nil
	case svc.PausePending:
		fallthrough
	case svc.Paused:
		fallthrough
	case svc.ContinuePending:
		fallthrough
	case svc.StopPending:
		fallthrough
	case svc.Stopped:
		return false, nil
	default:
		return false, fmt.Errorf("unknown status %v", status)
	}
}

func (c ServiceClient) InspectService(name string) (*Win32Service, error) {
	var wmiServices []Win32Service
	query := wmi.CreateQuery(&wmiServices, fmt.Sprintf("WHERE Name='%s'", name), "Win32_Service")

	if err := wmi.Query(query, &wmiServices); err != nil {
		return nil, fmt.Errorf("failed to inspect service by name %s: %v", name, err)
	} else if len(wmiServices) == 0 {
		return nil, nil
	} else {
		return &wmiServices[0], nil
	}
}

func (c ServiceClient) GetServiceStats(name string) (*wmiProcessStats, error) {
	stats := &wmiProcessStats{
		WorkingSetPrivate: 0,
		KernelModeTime:    0,
		UserModeTime:      0,
	}

	isRunning, err := c.IsServiceRunning(name)
	if err != nil || !isRunning {
		return stats, err
	}

	var wmiServices []Win32Service
	svcQuery := wmi.CreateQuery(&wmiServices, fmt.Sprintf("WHERE Name='%s'", name), "Win32_Service")

	if err := wmi.Query(svcQuery, &wmiServices); err != nil || len(wmiServices) == 0 {
		return stats, fmt.Errorf("failed to retrieve service by name %s: %v", name, err)
	}
	processId := wmiServices[0].ProcessId

	var win32Processes []win32Process
	query := wmi.CreateQuery(&win32Processes, fmt.Sprintf("WHERE ProcessID=%d", processId), "Win32_Process")

	if err := wmi.Query(query, &win32Processes); err != nil || len(win32Processes) == 0 {
		return stats, err
	}

	stats.KernelModeTime = win32Processes[0].KernelModeTime
	stats.UserModeTime = win32Processes[0].UserModeTime

	var formattedProcess []win32PerfFormattedDataPerfProcProcess

	// Query WMI for memory stats with the given process ids
	// We are only using the WorkingSetPrivate for our memory to better align the Windows Task Manager and the RSS field nomad is expecting
	query = wmi.CreateQuery(&formattedProcess, fmt.Sprintf("WHERE IDProcess=%d", processId), "Win32_PerfFormattedData_PerfProc_Process")
	if err := wmi.Query(query, &formattedProcess); err != nil || len(formattedProcess) == 0 {
		return stats, err
	}

	stats.WorkingSetPrivate = formattedProcess[0].WorkingSetPrivate

	// Need to multiply cpu stats by one hundred to align with nomad method CpuStats.Percent's expected decimal placement
	stats.KernelModeTime *= 100
	stats.UserModeTime *= 100

	return stats, nil
}
