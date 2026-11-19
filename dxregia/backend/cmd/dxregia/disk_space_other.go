//go:build !linux && !windows

package main

func diskSpace(string) (uint64, uint64, bool) { return 0, 0, false }
