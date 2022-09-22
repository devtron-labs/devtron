package sync

import "sync"

type KeyLock interface {
	Lock(key string)
	Unlock(key string)
	RLock(key string)
	RUnlock(key string)
}

type keyLock struct {
	guard sync.RWMutex
	locks map[string]*sync.RWMutex
}

func NewKeyLock() KeyLock {
	return &keyLock{
		guard: sync.RWMutex{},
		locks: map[string]*sync.RWMutex{},
	}
}

func (l *keyLock) getLock(key string) *sync.RWMutex {
	l.guard.RLock()
	if lock, ok := l.locks[key]; ok {
		l.guard.RUnlock()
		return lock
	}

	l.guard.RUnlock()
	l.guard.Lock()

	if lock, ok := l.locks[key]; ok {
		l.guard.Unlock()
		return lock
	}

	lock := &sync.RWMutex{}
	l.locks[key] = lock
	l.guard.Unlock()
	return lock
}

func (l *keyLock) Lock(key string) {
	l.getLock(key).Lock()
}

func (l *keyLock) Unlock(key string) {
	l.getLock(key).Unlock()
}

func (l *keyLock) RLock(key string) {
	l.getLock(key).RLock()
}

func (l *keyLock) RUnlock(key string) {
	l.getLock(key).RUnlock()
}
