/*
Package znet provides a lightweight and high-performance HTTP web framework.

It features a flexible router with middleware support, context-based request handling,
and various utilities for building web applications and RESTful APIs.

Basic Usage:

	package main

	import (
	    "github.com/sohaha/zlsgo/znet"
	)

	func main() {
	    r := znet.New() // Create a new router instance

	    r.SetMode(znet.DebugMode) // Enable debug mode for development

	    // Define a simple route
	    r.GET("/", func(c znet.Context) {
	        c.String(200, "hello world")
	    })

	    // Define a route with path parameters
	    r.GET("/users/:id", func(c znet.Context) {
	        id := c.Param("id")
	        c.JSON(200, map[string]interface{}{
	            "id": id,
	            "message": "User details",
	        })
	    })

	    // Start the HTTP server
	    znet.Run()
	}

The framework provides the following key features:

1. Routing: Support for RESTful routes with path parameters and wildcards
2. Middleware: Request processing pipeline with before and after handlers
3. Context: Request-scoped context with helper methods for request/response handling
4. Rendering: Support for various response formats (JSON, HTML, etc.)
5. Validation: Request data validation with customizable rules
6. Plugins: Extensions for common web features (CORS, GZIP, Rate Limiting, etc.)
*/
package znet
