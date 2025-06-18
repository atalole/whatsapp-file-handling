package utils

import (
	"fmt"
	"runtime"
)

func LogMemoryUsage() map[string]string {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	memoryUsage := map[string]string{
		"RSS (Resident Set Size)": fmt.Sprintf("%.2f MB", float64(memStats.Sys)/1024/1024),
		"Heap Total":              fmt.Sprintf("%.2f MB", float64(memStats.HeapSys)/1024/1024),
		"Heap Used":               fmt.Sprintf("%.2f MB", float64(memStats.HeapAlloc)/1024/1024),
		"External":                fmt.Sprintf("%.2f MB", float64(memStats.HeapReleased)/1024/1024),
		// Go does not directly expose CPU usage like Node.js, so these are placeholders
		"User":   "N/A",
		"System": "N/A",
	}

	return memoryUsage
}
