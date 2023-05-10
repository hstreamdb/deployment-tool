package spec

import (
	"fmt"
	"reflect"
)

// GetContainerCfg return the "ContainerCfg" filed value
func GetContainerCfg(i interface{}) ContainerCfg {
	if v := reflect.ValueOf(i).FieldByName("ContainerCfg"); v.IsValid() {
		return v.Interface().(ContainerCfg)
	}
	return ContainerCfg{}
}

// MergeContainerCfg merge two config, use rhs update lhs, and return updated lhs finally.
func MergeContainerCfg(lhs, rhs ContainerCfg) ContainerCfg {
	if len(rhs.Cpu) != 0 {
		lhs.Cpu = rhs.Cpu
	}
	if len(rhs.Memory) != 0 {
		lhs.Memory = rhs.Memory
	}
	// the default value of RemoveWhenExit is false, when rhs.RemoveWhenExit == true,
	// update lhs with rhs
	if rhs.RemoveWhenExit {
		lhs.RemoveWhenExit = rhs.RemoveWhenExit
	}
	// the default value of DisableRestart is false, when rhs.DisableRestart == true,
	// update lhs with rhs
	if rhs.DisableRestart {
		lhs.DisableRestart = rhs.DisableRestart
	}

	if len(rhs.Options) != 0 {
		lhs.Options = rhs.Options
	}
	return lhs
}

type MountPoints struct {
	Local  string
	Remote string
}

func GetDockerExecCmd(globalContainerSpec, serviceContainerSpec ContainerCfg, containerName string,
	hostMode bool, mountPoints ...MountPoints) []string {
	args := []string{"docker run -d"}
	if hostMode {
		args = append(args, "--network host")
	}
	args = append(args, fmt.Sprintf("--name %s", containerName))
	containerCfg := MergeContainerCfg(globalContainerSpec, serviceContainerSpec)
	args = append(args, containerCfg.GetCmd())
	if len(mountPoints) != 0 {
		for _, pair := range mountPoints {
			args = append(args, fmt.Sprintf("-v %s:%s", pair.Local, pair.Remote))
		}
	}
	return args
}
