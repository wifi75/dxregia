//go:build windows

package main

import (
	"context"
	"os"
)

// Su Windows il motore resta un processo figlio del servizio e si ferma con lui.

func keepEngineOnStop(string) bool { return false }

func adoptRunningEngine(string, string, func() bool) (*engineProc, bool) { return nil, false }

func startEngineProcess(service context.Context, perl, root string, logs *os.File) (*engineProc, error) {
	cmd := mojoCommand(service, perl, root, "cluster.pl", logs)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	return &engineProc{
		pid: int64(cmd.Process.Pid), done: done,
		term: func() {}, kill: func() { _ = cmd.Process.Kill() },
	}, nil
}
