package main

import (
	"testing"

	"github.com/akaritrading/platform/pkg/engine"
)

func TestBestNode(t *testing.T) {

	var stats = map[string]engine.MachineStat{
		"1": {CpuUsedPercent: 90.0, MemoryUsedPercent: 90.0},
		"2": {CpuUsedPercent: 30.0, MemoryUsedPercent: 90.0},
		"3": {CpuUsedPercent: 30.0, MemoryUsedPercent: 30.0},
		"4": {CpuUsedPercent: 90.0, MemoryUsedPercent: 30.0},
	}

	if ip, _ := bestNode(stats); ip != "3" {
		t.Fatal()
	}
}
