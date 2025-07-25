package zpprof

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/sohaha/zlsgo/zfile"
)

// SystemInfo SystemInfo
type SystemInfo struct {
	ServerName   string
	Runtime      string // runtime duration
	GoroutineNum string // goroutine count
	CPUNum       string // cpu core count
	UsedMem      string // current memory usage
	TotalMem     string // total allocated memory
	SysMem       string // system memory usage
	Lookups      string // pointer lookup count
	Mallocs      string // memory allocation count
	Frees        string // memory release count
	LastGCTime   string // time since last GC
	NextGC       string // next GC memory reclaim amount
	PauseTotalNs string // total GC pause time
	PauseNs      string // last GC pause time
	HeapInuse    string // heap memory in use
}

func NewSystemInfo(startTime time.Time) *SystemInfo {
	var afterLastGC string
	mstat := &runtime.MemStats{}
	runtime.ReadMemStats(mstat)
	costTime := int(time.Since(startTime).Seconds())
	if mstat.LastGC != 0 {
		afterLastGC = fmt.Sprintf("%.1fs", float64(time.Now().UnixNano()-int64(mstat.LastGC))/1000/1000/1000)
	} else {
		afterLastGC = "0"
	}

	serverName, _ := os.Hostname()

	return &SystemInfo{
		ServerName:   serverName,
		Runtime:      fmt.Sprintf("%d天%d小时%d分%d秒", costTime/(3600*24), costTime%(3600*24)/3600, costTime%3600/60, costTime%(60)),
		GoroutineNum: strconv.Itoa(runtime.NumGoroutine()),
		CPUNum:       strconv.Itoa(runtime.NumCPU()),
		HeapInuse:    zfile.SizeFormat(uint64(mstat.HeapInuse)),
		UsedMem:      zfile.SizeFormat(uint64(mstat.Alloc)),
		TotalMem:     zfile.SizeFormat(uint64(mstat.TotalAlloc)),
		SysMem:       zfile.SizeFormat(uint64(mstat.Sys)),
		Lookups:      strconv.FormatUint(mstat.Lookups, 10),
		Mallocs:      strconv.FormatUint(mstat.Mallocs, 10),
		Frees:        strconv.FormatUint(mstat.Frees, 10),
		LastGCTime:   afterLastGC,
		NextGC:       zfile.SizeFormat(uint64(mstat.NextGC)),
		PauseTotalNs: fmt.Sprintf("%.3fs", float64(mstat.PauseTotalNs)/1000/1000/1000),
		PauseNs:      fmt.Sprintf("%.3fs", float64(mstat.PauseNs[(mstat.NumGC+255)%256])/1000/1000/1000),
	}
}
