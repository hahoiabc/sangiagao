package handler

import (
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/pkg/cache"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

var startTime = time.Now()

type SystemHandler struct {
	cache cache.Cache
}

func NewSystemHandler(appCache cache.Cache) *SystemHandler {
	return &SystemHandler{cache: appCache}
}

type SystemStats struct {
	// Server
	Uptime    string `json:"uptime"`
	GoVersion string `json:"go_version"`
	Hostname  string `json:"hostname"`

	// CPU
	CPUCores   int     `json:"cpu_cores"`
	CPUPercent float64 `json:"cpu_percent"`

	// Memory
	MemTotal     uint64  `json:"mem_total_mb"`
	MemUsed      uint64  `json:"mem_used_mb"`
	MemPercent   float64 `json:"mem_percent"`

	// Disk
	DiskTotal    uint64  `json:"disk_total_gb"`
	DiskUsed     uint64  `json:"disk_used_gb"`
	DiskPercent  float64 `json:"disk_percent"`

	// Go runtime
	Goroutines   int    `json:"goroutines"`
	HeapAllocMB  uint64 `json:"heap_alloc_mb"`
	HeapSysMB    uint64 `json:"heap_sys_mb"`
	GCCycles     uint32 `json:"gc_cycles"`

	// Online users
	OnlineUsers  int      `json:"online_users"`
	OnlineIDs    []string `json:"online_ids"`
}

func (h *SystemHandler) GetStats(c *gin.Context) {
	stats := SystemStats{
		Uptime:     time.Since(startTime).Round(time.Second).String(),
		GoVersion:  runtime.Version(),
		Goroutines: runtime.NumGoroutine(),
		CPUCores:   runtime.NumCPU(),
	}

	hostname, _ := os.Hostname()
	stats.Hostname = hostname

	// Memory stats (Go runtime)
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	stats.HeapAllocMB = m.HeapAlloc / 1024 / 1024
	stats.HeapSysMB = m.HeapSys / 1024 / 1024
	stats.GCCycles = m.NumGC

	// System CPU
	cpuPercent, err := cpu.Percent(200*time.Millisecond, false)
	if err == nil && len(cpuPercent) > 0 {
		stats.CPUPercent = cpuPercent[0]
	}

	// System Memory
	vmStat, err := mem.VirtualMemory()
	if err == nil {
		stats.MemTotal = vmStat.Total / 1024 / 1024
		stats.MemUsed = vmStat.Used / 1024 / 1024
		stats.MemPercent = vmStat.UsedPercent
	}

	// Disk
	diskStat, err := disk.Usage("/")
	if err == nil {
		stats.DiskTotal = diskStat.Total / 1024 / 1024 / 1024
		stats.DiskUsed = diskStat.Used / 1024 / 1024 / 1024
		stats.DiskPercent = diskStat.UsedPercent
	}

	// Online users from Redis
	if h.cache != nil {
		keys, err := h.cache.KeysByPrefix(c.Request.Context(), "online:")
		if err == nil {
			stats.OnlineUsers = len(keys)
			ids := make([]string, len(keys))
			for i, k := range keys {
				ids[i] = strings.TrimPrefix(k, "online:")
			}
			stats.OnlineIDs = ids
		}
	}

	c.JSON(http.StatusOK, stats)
}
