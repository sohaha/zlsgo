package zsync_test

import (
	"context"
	"errors"
	"testing"
	"time"

	zls "github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zsync"
)

func TestNewPromise(t *testing.T) {
	tt := zls.NewTest(t)

	i := 0
	p := zsync.NewPromise(func() (int, error) {
		return 2, nil
	}).Finally(func() {
		i++
	})

	p = p.Then(func(i int) (int, error) {
		return i * 2, nil
	})

	res, err := p.Done()
	tt.NoError(err)
	tt.Equal(4, res)

	p = p.Then(func(i int) (int, error) {
		return i * 2, errors.New("this is an error")
	}).
		Finally(func() {
			i++
		}).
		Then(func(i int) (int, error) {
			return i * 2, errors.New("this is an new error")
		}).
		Finally(func() {
			i++
		}).
		Catch(func(err error) (int, error) {
			return 0, errors.New("catch: " + err.Error())
		})

	res, err = p.Done()
	if err == nil {
		t.Fatal("expected error")
	}
	tt.Equal("catch: this is an error", err.Error())
	tt.Equal(0, res)

	res, err = p.Finally(func() {
		i++
	}).
		Catch(func(err error) (int, error) {
			return 10, nil
		}).
		Then(func(i int) (int, error) {
			return i + 2, nil
		}).Done()
	tt.NoError(err)
	tt.Equal(12, res)

	tt.Equal(i, 4)
}

func TestNewPromiseContext(t *testing.T) {
	tt := zls.NewTest(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second/5)
	defer cancel()

	p1 := zsync.NewPromiseContext(ctx, func() (int, error) {
		time.Sleep(time.Second / 2)
		return 2, nil
	})

	var wg zsync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Go(func() {
			val, err := p1.Done()
			tt.EqualTrue(err != nil)
			tt.Equal(0, val)
			tt.Equal(context.Canceled, err)
		})
	}
	wg.Wait()

	p2 := zsync.NewPromiseContext(context.Background(), func() (int, error) {
		time.Sleep(time.Second / 2)
		return 2, nil
	})
	p3 := p2.Then(func(i int) (int, error) {
		return i * 2, nil
	})

	val, err := p3.Done()
	tt.Log(val, err)
	tt.NoError(err)
	tt.Equal(4, val)

	p3 = p3.Then(func(i int) (int, error) {
		return i * 5, nil
	})

	val, err = p3.Done()
	tt.Log(val, err)
	tt.NoError(err)
	tt.Equal(20, val)

	val, err = p2.Done()
	tt.Log(val, err)
	tt.NoError(err)
	tt.Equal(2, val)
}

func TestPromiseAll(t *testing.T) {
	tt := zls.NewTest(t)

	p1 := zsync.NewPromise(func() (int, error) {
		time.Sleep(time.Second / 5)
		return 2, nil
	})

	p2 := zsync.NewPromise(func() (int, error) {
		time.Sleep(time.Second / 2)
		return 20, nil
	})

	p3 := zsync.NewPromise(func() (int, error) {
		time.Sleep(time.Second)
		return 30, nil
	})

	all, err := zsync.PromiseAll(p1, p2, p3).Done()
	tt.NoError(err)
	tt.Equal([]int{2, 20, 30}, all)

	p4 := zsync.NewPromise(func() (int, error) {
		return 80, errors.New("this is an error")
	})

	p5 := zsync.NewPromise(func() (int, error) {
		return 100, nil
	})

	all, err = zsync.PromiseAll(p1, p4, p5).Done()
	tt.EqualTrue(err != nil)
	tt.Equal(0, len(all))
}

func TestPromiseRace(t *testing.T) {
	tt := zls.NewTest(t)

	p1 := zsync.NewPromise(func() (int, error) {
		time.Sleep(time.Second / 2)
		return 2, nil
	})

	p2 := zsync.NewPromise(func() (int, error) {
		time.Sleep(time.Second / 5)
		return 20, nil
	})

	p3 := zsync.NewPromise(func() (int, error) {
		time.Sleep(time.Second)
		return 30, nil
	})

	val, err := zsync.PromiseRace(p1, p2, p3).Done()
	tt.NoError(err)
	tt.Equal(20, val)

	p4 := zsync.NewPromise(func() (int, error) {
		return 80, errors.New("this is an error")
	})

	p5 := zsync.NewPromise(func() (int, error) {
		return 100, errors.New("p5 error")
	})

	val, err = zsync.PromiseRace(p3, p4, p5).Done()
	tt.EqualTrue(err != nil)
	tt.Equal(0, val)
}

func TestPromiseAny(t *testing.T) {
	tt := zls.NewTest(t)

	p1 := zsync.NewPromise(func() (int, error) {
		time.Sleep(time.Second / 2)
		return 2, nil
	})

	p2 := zsync.NewPromise(func() (int, error) {
		time.Sleep(time.Second / 5)
		return 20, nil
	})

	p3 := zsync.NewPromise(func() (int, error) {
		time.Sleep(time.Second)
		return 30, nil
	})

	val, err := zsync.PromiseAny(p1, p2, p3).Done()
	tt.NoError(err)
	tt.Equal(20, val)

	p4 := zsync.NewPromise(func() (int, error) {
		return 80, errors.New("this is an error")
	})

	p5 := zsync.NewPromise(func() (int, error) {
		return 100, errors.New("p5 error")
	})

	val, err = zsync.PromiseAny(p3, p4, p5).Done()
	tt.NoError(err)
	tt.Equal(30, val)

	val, err = zsync.PromiseAny(p4, p5).Done()
	tt.EqualTrue(err != nil)
	tt.Equal(0, val)
	tt.Log(err)
}
