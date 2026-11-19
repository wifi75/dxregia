package main

import "golang.org/x/sys/windows"

func diskSpace(path string) (free, total uint64, ok bool) {
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, 0, false
	}
	ok = windows.GetDiskFreeSpaceEx(name, &free, &total, nil) == nil
	return
}
