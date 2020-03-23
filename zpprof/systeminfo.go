/*
 * @Author: seekwe
 * @Date:   2019-05-09 16:09:12
 * @Last Modified by:   seekwe
 * @Last Modified time: 2019-05-28 15:22:30
 */

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
	Runtime      string // 运行时间
	GoroutineNum string // goroutine数量
	CPUNum       string // cpu核数
	UsedMem      string // 当前内存使用量
	TotalMem     string // 总分配的内存
	SysMem       string // 系统内存占用量
	Lookups      string // 指针查找次数
	Mallocs      string // 内存分配次数
	Frees        string // 内存释放次数
	LastGCTime   string // 距离上次GC时间
	NextGC       string // 下次GC内存回收量
	PauseTotalNs string // GC暂停时间总量
	PauseNs      string // 上次GC暂停时间
	HeapInuse    string // 正在使用的堆内存
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
		HeapInuse:    zfile.FileSizeFormat(uint64(mstat.HeapInuse)),
		UsedMem:      zfile.FileSizeFormat(uint64(mstat.Alloc)),
		TotalMem:     zfile.FileSizeFormat(uint64(mstat.TotalAlloc)),
		SysMem:       zfile.FileSizeFormat(uint64(mstat.Sys)),
		Lookups:      strconv.FormatUint(mstat.Lookups, 10),
		Mallocs:      strconv.FormatUint(mstat.Mallocs, 10),
		Frees:        strconv.FormatUint(mstat.Frees, 10),
		LastGCTime:   afterLastGC,
		NextGC:       zfile.FileSizeFormat(uint64(mstat.NextGC)),
		PauseTotalNs: fmt.Sprintf("%.3fs", float64(mstat.PauseTotalNs)/1000/1000/1000),
		PauseNs:      fmt.Sprintf("%.3fs", float64(mstat.PauseNs[(mstat.NumGC+255)%256])/1000/1000/1000),
	}
}
