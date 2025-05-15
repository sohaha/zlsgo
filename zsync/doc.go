/*
Package zsync provides enhanced synchronization primitives for concurrent programming in Go.

The package extends the standard library's sync package with additional features
and optimizations for common concurrency patterns. Key components include:

  - WaitGroup: An extended wait group with error handling and concurrency limiting
  - RBMutex: A reader-biased reader/writer mutual exclusion lock optimized for read-heavy workloads
  - Promise: A Go implementation of the Promise pattern for asynchronous operations
  - Context utilities: Tools for working with and combining multiple contexts

These primitives are designed to simplify concurrent programming while maintaining
high performance and safety.

Example usage of WaitGroup with concurrency limiting:

	wg := zsync.NewWaitGroup(10) // Limit to 10 concurrent goroutines
	for i := 0; i < 100; i++ {
		wg.Go(func() {
			// This work will be limited to 10 concurrent executions
			// ...
		})
	}
	err := wg.Wait() // Wait for all goroutines to complete

Example usage of Promise:

	p := zsync.NewPromise(func() (string, error) {
		// Perform async work
		return "result", nil
	})

	// Chain promises
	p2 := p.Then(func(result string) (string, error) {
		return result + " processed", nil
	})

	// Wait for the result
	result, err := p2.Done()
*/
package zsync
