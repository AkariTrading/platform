package engine

import "time"

type MachineStat struct {
	MemoryUsedPercent float64
	CpuUsedPercent    float64
	UpdatedAt         time.Time
}

var MachineStatsRedisKey = "node_stats"
