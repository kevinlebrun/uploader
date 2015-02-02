package uploader

import "sync"

type File interface {
	Upload()
	Key() string
}

type Uploader struct {
	files chan File
	kill  chan struct{}
	wg    sync.WaitGroup

	sync.Mutex
	running map[string]bool
}

func NewUploader(size int) *Uploader {
	s := &Uploader{
		files:   make(chan File, 128),
		kill:    make(chan struct{}),
		running: make(map[string]bool),
	}
	s.launch(size)
	return s
}

func (s *Uploader) launch(size int) {
	for i := 0; i < size; i++ {
		s.wg.Add(1)
		go s.worker()
	}
}

func (s *Uploader) worker() {
	defer s.wg.Done()

	for {
		select {
		case file, ok := <-s.files:
			if !ok {
				return
			}
			file.Upload()
			s.Lock()
			delete(s.running, file.Key())
			s.Unlock()
		case <-s.kill:
			return
		}
	}
}

func (s *Uploader) Upload(file File) bool {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.running[file.Key()]; !ok {
		s.files <- file
		s.running[file.Key()] = true
		return true
	}
	return false
}

func (s *Uploader) Wait() {
	s.wg.Wait()
}

func (s *Uploader) Close() {
	close(s.files)
}
