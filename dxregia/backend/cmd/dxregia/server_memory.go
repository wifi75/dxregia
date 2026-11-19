package main

import (
	"os"
	"strconv"
	"strings"
)

func serverMemory() (available, total uint64, ok bool) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, 0, false
	}
	return meminfo(string(data))
}

func meminfo(data string) (available, total uint64, ok bool) {
	foundAvailable := false
	for _, line := range strings.Split(data, "\n") {
		fields := strings.Fields(line)
		if len(fields) != 3 || fields[2] != "kB" {
			continue
		}
		if fields[0] != "MemTotal:" && fields[0] != "MemAvailable:" {
			continue
		}
		value, err := strconv.ParseUint(fields[1], 10, 54)
		if err != nil {
			continue
		}
		if fields[0] == "MemTotal:" {
			total = value * 1024
		} else {
			available, foundAvailable = value*1024, true
		}
	}
	return available, total, foundAvailable && total > 0 && available <= total
}
