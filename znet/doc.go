/*
Package znet provides web service
    package main

    import (
    	"github.com/sohaha/zlsgo/znet"
    )


    func main(){
    	r := znet.New()

    	r.SetMode(znet.DebugMode)
    	r.GET("/", func(c znet.Context) {
    		c.String(200, "hello world")
    	})

    	znet.Run()
    }
*/
package znet
