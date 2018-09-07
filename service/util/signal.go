package util

import (
	"sync"
	"time"
	"errors"
)

type Signal struct {
	isSet     bool
	autoReset bool
	*sync.Mutex
	chNotSet  chan interface{} // channel is open, when signal isn't set
}

func NewSignal(isSet, autoReset bool) (s *Signal) {
	s = &Signal{
		Mutex:     &sync.Mutex{},
		isSet:     isSet,
		autoReset: autoReset,
		chNotSet:  make(chan interface{}, 0),
	}
	if s.isSet {
		s.Lock()
		defer s.Unlock()
		s.isSet = true
		close(s.chNotSet)
	}
	return
}

func (s *Signal) Set() {
	// if signaled, the channel has to be closed
	// we can't test if the channel is already closed (without waiting with select), so we keep track of the state
	// in isSet, to avoid closing multiple times
	s.Lock()
	defer s.Unlock()
	if s.isSet {
		return
	}
	s.isSet = true
	close(s.chNotSet)
	return
}

func (s *Signal) Reset() {
	// in reset state, the channel has to exist, but mustn't be recreated if already existing (already in unset state)
	s.Lock()
	defer s.Unlock()
	if s.isSet {
		// channel shouldn't exist
		s.chNotSet = make(chan interface{}, 0)
		s.isSet = false
	}
	return
}

func (s Signal) IsSet() bool {
	return s.isSet
}

func (s *Signal) Wait() {
	select {
	case <-s.chNotSet: // when channel isn't closed, this blocks
		// if autoReset, recreate channel (setting signal to off)
		if s.autoReset {
			s.Lock()
			s.chNotSet = make(chan interface{}, 0)
			s.isSet = false
			s.Unlock()
		}
	}
	return
}

func (s *Signal) WaitTimeout(timeout time.Duration) error {
	select {
	case <-s.chNotSet: // when channel isn't closed, this blocks
		// if autoReset, recreate channel (setting signal to off)
		if s.autoReset {
			s.Lock()
			s.chNotSet = make(chan interface{}, 0)
			s.isSet = false
			s.Unlock()
		}
		return nil
	case <-time.After(timeout):
		return errors.New("Timeout reached")
	}
}
