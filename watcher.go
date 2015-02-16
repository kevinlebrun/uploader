package main

import (
	"os"
	"path/filepath"
	"sync"
	"time"
)

var watcher *Watcher

type Watcher struct {
	sync.Mutex
	Files map[string]struct{}
	pool  *Pool
}

func NewWatcher(pool *Pool) *Watcher {
	return &Watcher{Files: make(map[string]struct{}), pool: pool}
}

func (w *Watcher) Watch(dir string, pollInterval time.Duration, f func(path string) Job) {
	go w.watchFiles(dir, pollInterval, f)
	go w.release(w.pool.Done)
	w.pool.Wait()
}

func (w *Watcher) watchFiles(dir string, poll time.Duration, f func(string) Job) {
	for {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			create := false
			watcher.Lock()
			if _, ok := watcher.Files[path]; !ok && !info.IsDir() {
				watcher.Files[path] = struct{}{}
				create = true
			}
			watcher.Unlock()
			if create {
				w.pool.Exec(f(path))
			}
			return nil
		})
		time.Sleep(poll)
	}
}

func (w *Watcher) release(done <-chan Job) {
	for job := range done {
		watcher.Lock()
		delete(watcher.Files, job.Id())
		watcher.Unlock()
	}
}
