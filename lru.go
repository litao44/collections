package collections

import (
	"container/list"
	"sync"
)

type LRUInterface interface {
	Set(key string, val interface{})
	Get(key string) (interface{}, bool)
	Remove(key string)
	Len() int
}

var entityPool = sync.Pool{
	New: func() interface{} { return new(entity) },
}

type entity struct {
	key string
	val interface{}
}

type EvictCallback func(string, interface{})

var _ LRUInterface = &SimpleLRU{}

type SimpleLRU struct {
	cap           int
	evictCallback EvictCallback
	lookup        map[string]*list.Element
	data          *list.List
}

func NewSimpleLRU(cap int) *SimpleLRU {
	return NewSimpleLRUWithEvict(cap, nil)
}

func NewSimpleLRUWithEvict(cap int, callback EvictCallback) *SimpleLRU {
	if cap < 0 {
		panic("cap must be positive")
	}

	lru := &SimpleLRU{
		cap:           cap,
		evictCallback: callback,
		lookup:        make(map[string]*list.Element, cap),
		data:          list.New(),
	}

	return lru
}

func (l *SimpleLRU) Set(key string, val interface{}) {
	ele, exist := l.lookup[key]
	if exist {
		ele.Value.(*entity).val = val
		l.data.MoveToFront(ele)
	} else {
		e := entityPool.Get().(*entity)
		e.key = key
		e.val = val
		ele = l.data.PushFront(e)
	}

	l.lookup[key] = ele

	l.evict()
}

func (l *SimpleLRU) Get(key string) (interface{}, bool) {
	ele, ok := l.lookup[key]
	if !ok {
		return nil, false
	}

	l.data.MoveToFront(ele)
	return ele.Value.(*entity).val, true
}

func (l *SimpleLRU) Remove(key string) {
	if ele, ok := l.lookup[key]; ok {
		l.data.Remove(ele)
		delete(l.lookup, key)
	}
}

func (l *SimpleLRU) SimpleGet(key string) (interface{}, bool) {
	ele, ok := l.lookup[key]
	if !ok {
		return nil, false
	}

	return ele.Value.(*entity).val, true
}

func (l *SimpleLRU) Clear() {
	l.lookup = make(map[string]*list.Element, l.cap)
	l.data = list.New()
}

func (l *SimpleLRU) Len() int {
	return l.data.Len()
}

func (l *SimpleLRU) evict() {
	for l.Len() > l.cap {
		ele := l.data.Back()
		e := ele.Value.(*entity)
		if l.evictCallback != nil {
			l.evictCallback(e.key, e.val)
		}
		delete(l.lookup, e.key)
		l.data.Remove(ele)
	}
}

var _ LRUInterface = &LRU{}

type LRU struct {
	mu     sync.RWMutex
	simple *SimpleLRU
}

func NewLRU(cap int) *LRU {
	return NewLRUWithEvict(cap, nil)
}

func NewLRUWithEvict(cap int, callback EvictCallback) *LRU {
	return &LRU{
		mu:     sync.RWMutex{},
		simple: NewSimpleLRUWithEvict(cap, callback),
	}
}

func (l *LRU) Set(key string, val interface{}) {
	l.mu.Lock()
	l.simple.Set(key, val)
	l.mu.Unlock()
}

func (l *LRU) Get(key string) (interface{}, bool) {
	l.mu.Lock()
	val, ok := l.simple.Get(key)
	l.mu.Unlock()
	return val, ok
}

func (l *LRU) Len() int {
	return l.simple.Len()
}

func (l *LRU) Remove(key string) {
	l.mu.Lock()
	l.simple.Remove(key)
	l.mu.Unlock()
}
