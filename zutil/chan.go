package zutil

type (
	chanType int
	conf     struct {
		len *Uint32
		typ chanType
		cap int64
	}
	Opt         func(*conf)
	Chan[T any] struct {
		in, out chan T
		close   chan struct{}
		conf    conf
		q       []T
	}

	Option struct {
		Cap int
	}
)

const (
	unbuffered chanType = iota
	buffered
	unbounded
)

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

func (ch *Chan[T]) In() chan<- T { return ch.in }

func (ch *Chan[T]) Out() <-chan T { return ch.out }

func (ch *Chan[T]) Close() {
	switch ch.conf.typ {
	case buffered, unbuffered:
		close(ch.in)
		close(ch.close)
	default:
		ch.close <- struct{}{}
	}
}

func (ch *Chan[T]) Len() int {
	switch ch.conf.typ {
	case buffered, unbuffered:
		return len(ch.in)
	default:
		return int(ch.conf.len.Load()) + len(ch.in) + len(ch.out)
	}
}

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
