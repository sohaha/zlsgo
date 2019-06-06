/*
 * @Author: seekwe
 * @Date:   2019-05-10 13:52:27
 * @Last Modified by:   seekwe
 * @Last Modified time: 2019-05-28 15:24:24
 */

// Package zpprof provides a register for zweb framework to use net/http/pprof easily.
package zpprof

import (
	"github.com/sohaha/zlsgo/znet"
)

// Register Registration routing
func Register(r *znet.Engine, token string) (RouterGroup *znet.Engine) {

	// go tool pprof http://127.0.0.1:8081/debug/pprof/profile
	// go tool pprof -alloc_space http://127.0.0.1:8081/debug/pprof/heap
	// go tool pprof -inuse_space http://127.0.0.1:8081/debug/pprof/heap

	RouterGroup = r.Group("/debug", func(g *znet.Engine) {
		g.Use(authDebug(token))
		g.GET("", infoHandler)
		g.GET("/", redirectPprof)
		g.GET("/pprof", redirectPprof)
		g.GET("/pprof/", indexHandler)
		g.GET("/pprof/allocs", allocsHandler)
		g.GET("/pprof/mutex", mutexHandler)
		g.GET("/pprof/heap", heapHandler)
		g.GET("/pprof/goroutine", goroutineHandler)
		g.GET("/pprof/block", blockHandler)
		g.GET("/pprof/threadcreate", threadCreateHandler)
		g.GET("/pprof/cmdline", cmdlineHandler)
		g.GET("/pprof/profile", profileHandler)
		g.GET("/pprof/symbol", symbolHandler)
		g.POST("/pprof/symbol", symbolHandler)
		g.GET("/pprof/trace", traceHandler)
	})
	return
}
