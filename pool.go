package main

import "sync"

type Job interface {
	Execute()
	Id() string
}

type Pool struct {
	jobs chan Job
	Done chan Job
	kill chan struct{}
	wg   sync.WaitGroup
}

func NewPool(size int) *Pool {
	p := &Pool{
		jobs: make(chan Job, 128),
		Done: make(chan Job, 128),
		kill: make(chan struct{}),
	}
	p.launch(size)
	return p
}

func (p *Pool) launch(size int) {
	for i := 0; i < size; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

func (p *Pool) worker() {
	defer p.wg.Done()

	for {
		select {
		case job, ok := <-p.jobs:
			if !ok {
				return
			}
			job.Execute()
			p.Done <- job
		case <-p.kill:
			return
		}
	}
}

func (p *Pool) Exec(job Job) {
	p.jobs <- job
}

func (p *Pool) Wait() {
	p.wg.Wait()
}

func (p *Pool) Close() {
	close(p.jobs)
}
