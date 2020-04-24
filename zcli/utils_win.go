// +build windows

package zcli

import (
	"syscall"
	"unsafe"
)

func IsDoubleClickStartUp() bool {
	if name, err := GetParentProcessName(); err == nil && name == "explorer.exe" {
		// defer fmt.Scanln()
		return true
	}
	return false
}

func GetParentProcessName() (string, error) {
	snapshot, err := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return "", err
	}
	defer syscall.CloseHandle(snapshot)
	var procEntry syscall.ProcessEntry32
	procEntry.Size = uint32(unsafe.Sizeof(procEntry))
	if err = syscall.Process32First(snapshot, &procEntry); err != nil {
		return "", err
	}
	var (
		pid      = uint32(syscall.Getpid())
		pName    = make(map[uint32]string, 32)
		parentId = uint32(1<<32 - 1)
	)
	for {
		pName[procEntry.ProcessID] = syscall.UTF16ToString(procEntry.ExeFile[:])
		if procEntry.ProcessID == pid {
			parentId = procEntry.ParentProcessID
		}
		if s, ok := pName[parentId]; ok {
			return s, nil
		}
		err = syscall.Process32Next(snapshot, &procEntry)
		if err != nil {
			return "", err
		}
	}
}
