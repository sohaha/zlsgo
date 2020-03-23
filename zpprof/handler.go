/*
 * @Author: seekwe
 * @Date:   2019-05-10 14:00:48
 * @Last Modified by:   seekwe
 * @Last Modified time: 2019-06-05 11:44:30
 */

package zpprof

import (
	"fmt"
	"time"

	"net/http"
	"net/http/pprof"

	"github.com/sohaha/zlsgo/znet"
)

var startTime = time.Now()

func infoHandler(c *znet.Context) {
	m := NewSystemInfo(startTime)
	info := fmt.Sprintf("%s:%s\n", "服务器", m.ServerName)
	info += fmt.Sprintf("%s:%s\n", "运行时间", m.Runtime)
	info += fmt.Sprintf("%s:%s\n", "goroutine数量", m.GoroutineNum)
	info += fmt.Sprintf("%s:%s\n", "CPU核数", m.CPUNum)
	info += fmt.Sprintf("%s:%s\n", "当前内存使用量", m.UsedMem)
	info += fmt.Sprintf("%s:%s\n", "当前堆内存使用量", m.HeapInuse)
	info += fmt.Sprintf("%s:%s\n", "总分配的内存", m.TotalMem)
	info += fmt.Sprintf("%s:%s\n", "系统内存占用量", m.SysMem)
	info += fmt.Sprintf("%s:%s\n", "指针查找次数", m.Lookups)
	info += fmt.Sprintf("%s:%s\n", "内存分配次数", m.Mallocs)
	info += fmt.Sprintf("%s:%s\n", "内存释放次数", m.Frees)
	info += fmt.Sprintf("%s:%s\n", "距离上次GC时间", m.LastGCTime)
	info += fmt.Sprintf("%s:%s\n", "下次GC内存回收量", m.NextGC)
	info += fmt.Sprintf("%s:%s\n", "GC暂停时间总量", m.PauseTotalNs)
	info += fmt.Sprintf("%s:%s\n", "上次GC暂停时间", m.PauseNs)
	fmt.Fprint(c.Writer, info)
}

func indexHandler(c *znet.Context) {
	pprof.Index(c.Writer, c.Request)
}

func allocsHandler(c *znet.Context) {
	pprof.Handler("allocs").ServeHTTP(c.Writer, c.Request)
}

func mutexHandler(c *znet.Context) {
	pprof.Handler("mutex").ServeHTTP(c.Writer, c.Request)
}

func heapHandler(c *znet.Context) {
	pprof.Handler("heap").ServeHTTP(c.Writer, c.Request)
}

func goroutineHandler(c *znet.Context) {
	pprof.Handler("goroutine").ServeHTTP(c.Writer, c.Request)
}

func blockHandler(c *znet.Context) {
	pprof.Handler("block").ServeHTTP(c.Writer, c.Request)
}

func threadCreateHandler(c *znet.Context) {
	pprof.Handler("threadcreate").ServeHTTP(c.Writer, c.Request)
}

func cmdlineHandler(c *znet.Context) {
	pprof.Cmdline(c.Writer, c.Request)
}

func profileHandler(c *znet.Context) {
	pprof.Profile(c.Writer, c.Request)
}

func symbolHandler(c *znet.Context) {
	pprof.Symbol(c.Writer, c.Request)
}

func traceHandler(c *znet.Context) {
	pprof.Trace(c.Writer, c.Request)
}

func redirectPprof(c *znet.Context) {
	http.Redirect(c.Writer, c.Request, "/debug/pprof/", http.StatusFound)
}

func authDebug(token string) znet.HandlerFunc {
	return func(c *znet.Context) {
		getToken := c.DefaultQuery("token", c.GetCookie("debug-token"))
		c.SetCookie("debug-token", token, 600)
		if token == "" || getToken == token {
			c.Next()
		} else {
			c.Byte(401, []byte("No permission"))
		}
	}
}
