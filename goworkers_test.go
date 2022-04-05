package goworkers

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

const timeout = 5 * time.Second

func Example() {
	ch := make(chan int)

	pool := Init(func() {
		i := <-ch
		fmt.Println(i * i)
	})

	pool.Add(1)
	ch <- 2
	ch <- 3
	wait := pool.Stop()
	ch <- 4
	<-wait
	// Output: 4
	// 9
	// 16
}

func testWorker(c *int, mu *sync.Mutex, ch1, ch2, ch3 chan struct{}) func() {
	return func() {
		defer func() {
			mu.Lock()
			*c--
			mu.Unlock()
			ch3 <- struct{}{}
		}()

		mu.Lock()
		*c++
		mu.Unlock()

		ch1 <- struct{}{}
		<-ch2
	}
}

func clearChannel(ch chan struct{}) {
	for ok := true; ok; _, ok = <-ch {
	}
}

func TestStartAndStopWithoutWorker(t *testing.T) {
	t.Parallel()

	ch := make(chan struct{})

	pool := Init(func() {
		ch <- struct{}{}
	})

	select {
	case <-ch:
		t.Errorf("Unexpected running worker")
	case <-pool.Stop():
	case <-time.After(timeout):
		t.Errorf("Elapsed stop timeout")
	}
}

func TestRemoveWithoutWorker(t *testing.T) {
	t.Parallel()

	ch := make(chan struct{})

	pool := Init(func() {
		ch <- struct{}{}
	})

	pool.Remove(1)

	select {
	case <-ch:
		t.Errorf("Unexpected running worker")
	case <-pool.Stop():
	case <-time.After(timeout):
		t.Errorf("Elapsed stop timeout")
	}
}

func TestAddTenWorkers(t *testing.T) {
	t.Parallel()

	const n = 10
	c := 0
	mu := &sync.Mutex{}
	ch1 := make(chan struct{})
	ch2 := make(chan struct{})
	ch3 := make(chan struct{})

	go clearChannel(ch3)
	defer close(ch3)
	pool := Init(testWorker(&c, mu, ch1, ch2, ch3))

	pool.Add(10)
	for i := 0; i < n; i++ {
		<-ch1
	}
	for i := 0; i < n; i++ {
		ch2 <- struct{}{}
		<-ch1
	}

	mu.Lock()
	if c != n {
		t.Errorf("Get %d workers instead of %d", c, n)
	}
	mu.Unlock()

	close(ch2)
	go clearChannel(ch1)
	defer close(ch1)
	select {
	case <-pool.Stop():
	case <-time.After(timeout):
		t.Errorf("Elapsed stop timeout")
	}
}

func TestSetTenWorkers(t *testing.T) {
	t.Parallel()

	const n = 10
	c := 0
	mu := &sync.Mutex{}
	ch1 := make(chan struct{})
	ch2 := make(chan struct{})
	ch3 := make(chan struct{})

	go clearChannel(ch3)
	defer close(ch3)
	pool := Init(testWorker(&c, mu, ch1, ch2, ch3))

	pool.Set(n)
	for i := 0; i < n; i++ {
		<-ch1
	}
	for i := 0; i < n; i++ {
		ch2 <- struct{}{}
		<-ch1
	}

	mu.Lock()
	if c != n {
		t.Errorf("Get %d workers instead of %d", c, n)
	}
	mu.Unlock()

	close(ch2)
	go clearChannel(ch1)
	defer close(ch1)
	select {
	case <-pool.Stop():
	case <-time.After(timeout):
		t.Errorf("Elapsed stop timeout")
	}
}

func TestRemoveHalfWorkers(t *testing.T) {
	t.Parallel()

	const n = 10
	c := 0
	mu := &sync.Mutex{}
	ch1 := make(chan struct{})
	ch2 := make(chan struct{})
	ch3 := make(chan struct{})

	go clearChannel(ch3)
	defer close(ch3)
	pool := Init(testWorker(&c, mu, ch1, ch2, ch3))

	pool.Set(2 * n)
	for i := 0; i < 2*n; i++ {
		<-ch1
	}
	pool.Remove(n)
	for i := 0; i < n; i++ {
		ch2 <- struct{}{}
	}
	for i := 0; i < n; i++ {
		ch2 <- struct{}{}
		<-ch1
	}

	mu.Lock()
	if c != n {
		t.Errorf("Get %d workers instead of %d", c, n)
	}
	mu.Unlock()

	close(ch2)
	go clearChannel(ch1)
	defer close(ch1)
	select {
	case <-pool.Stop():
	case <-time.After(timeout):
		t.Errorf("Elapsed stop timeout")
	}
}

func TestPauseWorkers(t *testing.T) {
	t.Parallel()

	const n = 10
	c := 0
	mu := &sync.Mutex{}
	ch1 := make(chan struct{})
	ch2 := make(chan struct{})
	ch3 := make(chan struct{})

	close(ch2)
	pool := Init(testWorker(&c, mu, ch1, ch2, ch3))

	pool.Set(n)
	for i := 0; i < n; i++ {
		<-ch1
	}
	_ = pool.Pause()
	for i := 0; i < n; i++ {
		<-ch3
	}

	select {
	case <-ch1:
		t.Errorf("Not paused worker")
	default:
	}

	select {
	case <-pool.Stop():
	case <-time.After(timeout):
		t.Errorf("Elapsed stop timeout")
	}
}

func TestGetWorkers(t *testing.T) {
	t.Parallel()

	const n = 10
	ch := make(chan struct{})

	pool := Init(func() {
		<-ch
	})

	pool.Set(10)
	c := pool.Get()
	if c != n {
		t.Errorf("Get %d workers instead of %d", c, n)
	}

	close(ch)
	select {
	case <-pool.Stop():
	case <-time.After(timeout):
		t.Errorf("Elapsed stop timeout")
	}
}

func TestActiveWorkers(t *testing.T) {
	t.Parallel()

	const n = 10
	c := 0
	mu := &sync.Mutex{}
	ch1 := make(chan struct{})
	ch2 := make(chan struct{})
	ch3 := make(chan struct{})

	go clearChannel(ch3)
	defer close(ch3)
	pool := Init(testWorker(&c, mu, ch1, ch2, ch3))

	pool.Set(n)
	for i := 0; i < n; i++ {
		<-ch1
	}

	a := pool.Active()
	if c != int(a) {
		t.Errorf("Get %d workers instead of %d", c, a)
	}

	go clearChannel(ch1)
	defer close(ch1)
	close(ch2)
	select {
	case <-pool.Stop():
	case <-time.After(timeout):
		t.Errorf("Elapsed stop timeout")
	}
}

func TestStateWorkers(t *testing.T) {
	t.Parallel()

	const n = 10
	c := 0
	mu := &sync.Mutex{}
	ch1 := make(chan struct{})
	ch2 := make(chan struct{})
	ch3 := make(chan struct{})

	pool := Init(testWorker(&c, mu, ch1, ch2, ch3))

	pool.Set(3 * n)
	for i := 0; i < 3*n; i++ {
		<-ch1
	}
	pool.Set(n)
	for i := 0; i < n; i++ {
		ch2 <- struct{}{}
		<-ch3
	}

	mu.Lock()
	_, w := pool.State()
	if c != 2*n {
		t.Errorf("Get %d workers instead of %d", c, 2*n)
	}
	if w != n {
		t.Errorf("Get %d workers instead of %d", w, n)
	}
	mu.Unlock()

	go clearChannel(ch3)
	defer close(ch3)
	close(ch2)
	select {
	case <-pool.Stop():
	case <-time.After(timeout):
		t.Errorf("Elapsed stop timeout")
	}
}

func TestConcurrentWorkers(t *testing.T) {
	t.Parallel()

	const n = 10
	c := 0
	mu := &sync.Mutex{}
	ch1 := make(chan struct{})
	ch2 := make(chan struct{})
	ch3 := make(chan struct{})

	pool1 := Init(testWorker(&c, mu, ch1, ch2, ch3))
	pool2 := Init(testWorker(&c, mu, ch1, ch2, ch3))
	pool3 := Init(testWorker(&c, mu, ch1, ch2, ch3))
	go clearChannel(ch3)
	defer close(ch3)

	pool1.Set(1 * n)
	pool2.Set(2 * n)
	pool3.Set(3 * n)

	for i := 0; i < 6*n; i++ {
		<-ch1
	}

	var a uint
	a = pool1.Get()
	if a != 1*n {
		t.Errorf("Get %d workers instead of %d (pool 1)", a, 1*n)
	}
	a = pool2.Get()
	if a != 2*n {
		t.Errorf("Get %d workers instead of %d (pool 2)", a, 2*n)
	}
	a = pool3.Get()
	if a != 3*n {
		t.Errorf("Get %d workers instead of %d (pool 3)", a, 3*n)
	}

	go clearChannel(ch1)
	defer close(ch1)
	close(ch2)
	select {
	case <-pool1.Stop():
	case <-time.After(timeout):
		t.Errorf("Elapsed stop timeout (pool 1)")
	}
	select {
	case <-pool2.Stop():
	case <-time.After(timeout):
		t.Errorf("Elapsed stop timeout (pool 2)")
	}
	select {
	case <-pool3.Stop():
	case <-time.After(timeout):
		t.Errorf("Elapsed stop timeout (pool 3)")
	}
}
