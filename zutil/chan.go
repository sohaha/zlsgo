//go:build go1.18
// +build go1.18

package zutil

type (
	// chanType represents the type of channel behavior (unbuffered, buffered, or unbounded)
	chanType int

	// conf holds configuration for a Chan instance
	conf struct {
		len *Uint32  // Current length of the queue for unbounded channels
		typ chanType // Type of channel (unbuffered, buffered, or unbounded)
		cap int64    // Capacity for buffered channels, -1 for unbounded
	}

	// Opt is a function type for configuring a Chan instance
	Opt func(*conf)

	// Chan is a generic channel implementation that supports unbuffered, buffered,
	// and unbounded channel behaviors. Unbounded channels will never block on send
	// operations, dynamically growing their internal queue as needed.
	Chan[T any] struct {
		in, out chan T        // Input and output channels
		close   chan struct{} // Channel for signaling close operations
		conf    conf          // Configuration
		q       []T           // Internal queue for unbounded channels
	}

	// Options contains configuration options for channel creation
	Options struct {
		Cap int // Capacity of the channel
	}
)

const (
	// unbuffered represents a channel with no buffer (sends block until received)
	unbuffered chanType = iota
	// buffered represents a channel with a fixed-size buffer
	buffered
	// unbounded represents a channel with an unlimited buffer (sends never block)
	unbounded
)

// NewChan creates a new generic channel with the specified capacity.
// The behavior depends on the capacity parameter:
//   - cap == 0: Creates an unbuffered channel (sends block until received)
//   - cap > 0: Creates a buffered channel with the specified capacity
//   - cap < 0 or not provided: Creates an unbounded channel (sends never block)
func NewChan[T any](cap ...int) *Chan[T] {
	o := conf{
		typ: unbounded,
		cap: -1,
		len: NewUint32(0),
	}

	if len(cap) > 0 {
		if cap[0] == 0 {
			o.cap = int64(0)
			o.typ = unbuffered
		} else if cap[0] > 0 {
			o.cap = int64(cap[0])
			o.typ = buffered
		} else {
			o.cap = int64(-1)
			o.typ = unbounded
		}
	}

	ch := &Chan[T]{conf: o, close: make(chan struct{})}
	switch ch.conf.typ {
	case unbuffered:
		ch.in = make(chan T)
		ch.out = ch.in
	case buffered:
		ch.in = make(chan T, ch.conf.cap)
		ch.out = ch.in
	case unbounded:
		ch.in = make(chan T, 16)
		ch.out = make(chan T, 16)
		go ch.process()
	}
	return ch
}

// In returns the send-only channel for sending values.
// This is the channel that producers should use to send values.
func (ch *Chan[T]) In() chan<- T { return ch.in }

// Out returns the receive-only channel for receiving values.
// This is the channel that consumers should use to receive values.
func (ch *Chan[T]) Out() <-chan T { return ch.out }

// Close closes the channel, preventing further sends.
// For unbounded channels, this will drain the internal queue
// and ensure all sent values can still be received.
func (ch *Chan[T]) Close() {
	switch ch.conf.typ {
	case buffered, unbuffered:
		close(ch.in)
		close(ch.close)
	default:
		ch.close <- struct{}{}
	}
}

// Len returns the current number of elements in the channel.
// For unbounded channels, this includes elements in the internal queue
// as well as the input and output buffers.
func (ch *Chan[T]) Len() int {
	switch ch.conf.typ {
	case buffered, unbuffered:
		return len(ch.in)
	default:
		return int(ch.conf.len.Load()) + len(ch.in) + len(ch.out)
	}
}

// process is an internal goroutine that handles the unbounded channel behavior.
// It moves elements from the input channel to the internal queue and then to the output channel.
// This enables the unbounded behavior where sends never block.
func (ch *Chan[T]) process() {
	var nilT T

	ch.q = make([]T, 0, 1<<10)
	for {
		select {
		case e, ok := <-ch.in:
			if !ok {
				return
			}
			ch.conf.len.Add(1)
			ch.q = append(ch.q, e)
		case <-ch.close:
			ch.closeUnbounded()
			return
		}

		for len(ch.q) > 0 {
			select {
			case ch.out <- ch.q[0]:
				ch.conf.len.Sub(1)
				ch.q[0] = nilT
				ch.q = ch.q[1:]
			case e, ok := <-ch.in:
				if !ok {
					return
				}
				ch.conf.len.Add(1)
				ch.q = append(ch.q, e)
			case <-ch.close:
				ch.closeUnbounded()
				return
			}
		}
		if cap(ch.q) < 1<<5 {
			ch.q = make([]T, 0, 1<<10)
		}
	}
}

// closeUnbounded is an internal method that handles closing an unbounded channel.
// It drains any remaining elements from the input channel and internal queue,
// ensuring they are all sent to the output channel before closing it.
func (ch *Chan[T]) closeUnbounded() {
	var nilT T

	close(ch.in)

	for e := range ch.in {
		ch.q = append(ch.q, e)
	}

	for len(ch.q) > 0 {
		ch.out <- ch.q[0]
		ch.q[0] = nilT
		ch.q = ch.q[1:]
	}

	close(ch.out)
	close(ch.close)
}
