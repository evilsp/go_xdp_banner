package wait

import "sync"

// SingleInstance will make sure only one instance of the object is running
// at the same time.
type SingleInstance struct {
	lock sync.Mutex
	load bool

	NewInstance func() Instance
	instance    Instance
}

type Instance interface {
	Run()
	Stop()
}

func (s *SingleInstance) Run() {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.load {
		return
	}

	s.instance = s.NewInstance()
	if s.instance == nil {
		return
	}

	go s.instance.Run()
	s.load = true
}

func (s *SingleInstance) Stop() {
	s.lock.Lock()
	defer s.lock.Unlock()

	if !s.load {
		return
	}

	s.instance.Stop()
	s.load = false
	s.instance = nil
}
