// goworkers is a simple but effecient dynamic workers pool
package goworkers

// Pool is a workers pool
type Pool struct {
	add    chan struct{}
	remove chan struct{}
	finish chan struct{}
	stop   chan struct{}
	end    chan struct{}
	state  chan chan [2]uint
	set    chan uint
}

// Init returns a pool without worker with f function
func Init(f func()) *Pool {
	p := Pool{
		add:    make(chan struct{}),
		remove: make(chan struct{}),
		finish: make(chan struct{}),
		stop:   make(chan struct{}),
		end:    make(chan struct{}),
		state:  make(chan chan [2]uint),
		set:    make(chan uint),
	}

	go p.start(f)

	return &p
}

func (p *Pool) start(function func()) {
	active := uint(0)
	wished := uint(0)
	closed := false

	for {
		if closed && active < 1 {
			break
		}

		select {
		case ch := <-p.state:
			ch <- [2]uint{active, wished}
		case <-p.stop:
			wished = 0
			closed = true
		case want := <-p.set:
			if !closed {
				wished = want
			}
			for active < want {
				go p.launch(function)
				active++
			}
		case <-p.add:
			if !closed {
				wished++
			}
			go p.launch(function)
			active++
		case <-p.remove:
			if wished > 0 {
				wished--
			}
		case <-p.finish:
			active--
			for active < wished {
				go p.launch(function)
				active++
			}
		}
	}

	p.end <- struct{}{}
}

func (p *Pool) launch(function func()) {
	defer func() {
		p.finish <- struct{}{}
	}()

	function()
}

// Set n workers
func (p *Pool) Set(n uint) {
	p.set <- n
}

// Add n workers
func (p *Pool) Add(n uint) {
	for i := uint(0); i < n; i++ {
		p.add <- struct{}{}
	}
}

// Remove n workers (without killing them)
func (p *Pool) Remove(n uint) {
	for i := uint(0); i < n; i++ {
		p.remove <- struct{}{}
	}
}

// Pause remove all workers
func (p *Pool) Pause() uint {
	n := p.Get()
	p.Set(0)
	return n
}

// Stop the workers pool
func (p *Pool) Stop() chan struct{} {
	p.stop <- struct{}{}
	return p.end
}

// Get returns wished number of workers
func (p *Pool) Get() (wished uint) {
	_, wished = p.State()
	return wished
}

// Active returns current number of workers
func (p *Pool) Active() (active uint) {
	active, _ = p.State()
	return active
}

// State returns current and wished number of workers
func (p *Pool) State() (active uint, wished uint) {
	ch := make(chan [2]uint)
	p.state <- ch
	r := <-ch
	return r[0], r[1]
}
