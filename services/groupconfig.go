package services

import (
	"github.com/juju/errgo"
	"github.com/yext/edward/common"
)

var _ ServiceOrGroup = &ServiceGroupConfig{}

// ServiceGroupConfig is a group of services that can be managed together
type ServiceGroupConfig struct {
	// A name for this group, used to identify it in commands
	Name string
	// Full services contained within this group
	Services []*ServiceConfig
	// Groups on which this group depends
	Groups []*ServiceGroupConfig

	Logger common.Logger
}

func (c *ServiceGroupConfig) printf(format string, v ...interface{}) {
	if c.Logger == nil {
		return
	}
	c.Logger.Printf(format, v...)
}

func (sg *ServiceGroupConfig) GetName() string {
	return sg.Name
}

func (sg *ServiceGroupConfig) Build() error {
	println("Building group: ", sg.Name)
	for _, group := range sg.Groups {
		err := group.Build()
		if err != nil {
			return err
		}
	}
	for _, service := range sg.Services {
		err := service.Build()
		if err != nil {
			return err
		}
	}
	return nil
}

func (sg *ServiceGroupConfig) Start() error {
	println("Starting group:", sg.Name)
	for _, group := range sg.Groups {
		err := group.Start()
		if err != nil {
			// Always fail if any services in a dependant group failed
			return err
		}
	}
	var outErr error = nil
	for _, service := range sg.Services {
		err := service.Start()
		if err != nil {
			return err
		}
	}
	return outErr
}

func (sg *ServiceGroupConfig) Stop() error {
	println("=== Group:", sg.Name, "===")
	// TODO: Do this in reverse
	for _, service := range sg.Services {
		err := service.Stop()
		if err != nil {
			return errgo.Mask(err)
		}
	}
	for _, group := range sg.Groups {
		err := group.Stop()
		if err != nil {
			return errgo.Mask(err)
		}
	}
	return nil
}

func (sg *ServiceGroupConfig) Status() ([]ServiceStatus, error) {
	var outStatus []ServiceStatus
	for _, service := range sg.Services {
		statuses, err := service.Status()
		if err != nil {
			return outStatus, errgo.Mask(err)
		}
		outStatus = append(outStatus, statuses...)
	}
	for _, group := range sg.Groups {
		statuses, err := group.Status()
		if err != nil {
			return outStatus, errgo.Mask(err)
		}
		outStatus = append(outStatus, statuses...)
	}
	return outStatus, nil
}

func (sg *ServiceGroupConfig) IsSudo() bool {
	for _, service := range sg.Services {
		if service.IsSudo() {
			return true
		}
	}
	for _, group := range sg.Groups {
		if group.IsSudo() {
			return true
		}
	}

	return false
}

func (s *ServiceGroupConfig) GetWatchDirs() map[string]*ServiceConfig {
	watchMap := make(map[string]*ServiceConfig)
	for _, service := range s.Services {
		for watch, s := range service.GetWatchDirs() {
			watchMap[watch] = s
		}
	}
	for _, group := range s.Groups {
		for watch, s := range group.GetWatchDirs() {
			watchMap[watch] = s
		}
	}
	return watchMap
}
