// The MIT License (MIT)
//
// Copyright (c) 2020 Ryan Fowler
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package lru

import (
	"container/list"
	"context"
	"sync"

	"github.com/ryanfowler/distcache"
)

var _ distcache.Store = (*LRU)(nil)

type LRU struct {
	maxBytes int

	mu      sync.Mutex
	size    int
	valList *list.List
	values  map[string]*lruValue
}

type lruValue struct {
	key  string
	val  []byte
	elem *list.Element
}

func New(maxBytes int) *LRU {
	return &LRU{
		maxBytes: maxBytes,
		valList:  list.New(),
		values:   make(map[string]*lruValue),
	}
}

func (l *LRU) Get(ctx context.Context, key string) ([]byte, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if value, ok := l.values[key]; ok {
		l.valList.MoveToFront(value.elem)
		return value.val, nil
	}
	return nil, nil
}

func (l *LRU) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.values)
}

func (l *LRU) Size() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.size
}

func (l *LRU) Set(ctx context.Context, key string, val []byte) error {
	if val == nil {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if value, ok := l.values[key]; ok {
		l.size += len(val) - len(value.val)
		value.val = val
		l.valList.MoveToFront(value.elem)
	} else {
		l.size += len(key) + len(val)
		value = &lruValue{key: key, val: val}
		value.elem = l.valList.PushFront(value)
		l.values[key] = value
	}

	l.evict()
	return nil
}

func (l *LRU) evict() {
	for l.size > l.maxBytes {
		tail := l.valList.Back()
		if tail == nil {
			return
		}
		value := l.valList.Remove(tail).(*lruValue)
		delete(l.values, value.key)
		l.size -= (len(value.key) + len(value.val))
	}
}
