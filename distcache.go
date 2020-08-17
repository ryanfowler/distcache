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

package distcache

import (
	"context"
	"io"
	"math/rand"
	"sync"

	"golang.org/x/sync/singleflight"
)

type Getter interface {
	Get(ctx context.Context, key string) ([]byte, error)
}

type GetterFunc func(ctx context.Context, key string) ([]byte, error)

func (gf GetterFunc) Get(ctx context.Context, key string) ([]byte, error) {
	return gf(ctx, key)
}

type Setter interface {
	Set(ctx context.Context, key string, val []byte) error
}

type Peer interface {
	io.Closer
	Get(ctx context.Context, key string) ([]byte, ResultSource, error)
}

type PeerCreator interface {
	NewPeer(addr string) Peer
}

type Store interface {
	Getter
	Setter
}

type Cache struct {
	me string

	hotStore   Store
	localStore Store

	getter      Getter
	peerCreator PeerCreator

	single singleflight.Group

	mu    sync.Mutex
	hash  *peerHash
	peers map[string]Peer

	muSetPeers sync.Mutex
}

type Options struct {
	Me          string
	HotStore    Store
	LocalStore  Store
	Getter      Getter
	PeerCreator PeerCreator
	Peers       []string
}

func New(opts Options) *Cache {
	c := &Cache{
		me:          opts.Me,
		hotStore:    opts.HotStore,
		localStore:  opts.LocalStore,
		getter:      opts.Getter,
		peerCreator: opts.PeerCreator,
	}
	c.SetPeers(opts.Peers...)
	return c
}

type ResultSource int

const (
	ResultNone ResultSource = iota
	ResultHotCache
	ResultLocalCache
	ResultLocalGet
	ResultPeerCache
	ResultPeerGet
)

func (rs ResultSource) String() string {
	switch rs {
	case ResultNone:
		return "none"
	case ResultHotCache:
		return "cache_hot"
	case ResultLocalCache:
		return "cache_local"
	case ResultLocalGet:
		return "get_local"
	case ResultPeerCache:
		return "cache_peer"
	case ResultPeerGet:
		return "get_peer"
	}
	return "unknown"
}

func (c *Cache) Get(ctx context.Context, key string) ([]byte, ResultSource, error) {
	ch := c.single.DoChan(key, func() (interface{}, error) {
		return c.get(ctx, key)
	})
	var res singleflight.Result
	select {
	case <-ctx.Done():
		return nil, ResultNone, ctx.Err()
	case res = <-ch:
	}
	if res.Err != nil {
		return nil, ResultNone, res.Err
	}
	result := res.Val.(getResult)
	return result.Value, result.Source, nil
}

type getResult struct {
	Source ResultSource
	Value  []byte
}

func (c *Cache) get(ctx context.Context, key string) (getResult, error) {
	if res, ok := c.getFromStores(ctx, key); ok {
		return res, nil
	}

	c.mu.Lock()
	hash := c.hash
	peers := c.peers
	c.mu.Unlock()

	addr := hash.GetPeer([]byte(key))
	if addr == c.me {
		return c.getLocal(ctx, key)
	}

	if peer, ok := peers[addr]; ok {
		val, err := c.getFromPeer(ctx, peer, key)
		if err == nil {
			return val, nil
		}
		// TODO(ryanfowler): What to do with error here?
	}

	// Otherwise, fallback to getting locally.
	val, err := c.getter.Get(ctx, key)
	if err != nil {
		return getResult{}, err
	}
	return getResult{Source: ResultLocalGet, Value: val}, nil
}

func (c *Cache) getFromStores(ctx context.Context, key string) (getResult, bool) {
	// TODO(ryanfowler): How to handle errors here?
	val, err := c.hotStore.Get(ctx, key)
	if err == nil && val != nil {
		return getResult{Source: ResultHotCache, Value: val}, true
	}
	val, err = c.localStore.Get(ctx, key)
	if err == nil && val != nil {
		return getResult{Source: ResultLocalCache, Value: val}, true
	}
	return getResult{}, false
}

func (c *Cache) getFromPeer(ctx context.Context, peer Peer, key string) (getResult, error) {
	val, src, err := peer.Get(ctx, key)
	if err != nil {
		return getResult{}, err
	}
	c.populateHotStore(ctx, key, val)
	return getResult{Source: src, Value: val}, nil
}

func (c *Cache) getLocal(ctx context.Context, key string) (getResult, error) {
	val, err := c.getter.Get(ctx, key)
	if err != nil {
		return getResult{}, err
	}
	c.populateLocalStore(ctx, key, val)
	return getResult{Source: ResultLocalGet, Value: val}, nil
}

func (c *Cache) populateHotStore(ctx context.Context, key string, val []byte) {
	// TODO(ryanfowler): How do we populate the hot cache?
	if rand.Int31n(5) == 0 {
		_ = c.hotStore.Set(ctx, key, val)
	}
}

func (c *Cache) populateLocalStore(ctx context.Context, key string, val []byte) {
	// TODO(ryanfowler): Error handling?
	_ = c.localStore.Set(ctx, key, val)
}

func (c *Cache) SetPeers(peers ...string) {
	c.muSetPeers.Lock()
	defer c.muSetPeers.Unlock()

	c.mu.Lock()
	existingPeers := c.peers
	c.mu.Unlock()

	// Create new hash and peer map.
	newHash := newPeerHash(peers...)
	newPeers := make(map[string]Peer, len(peers))
	for _, addr := range peers {
		if addr == c.me {
			continue
		}
		if peer, ok := existingPeers[addr]; ok {
			newPeers[addr] = peer
			continue
		}
		newPeers[addr] = c.peerCreator.NewPeer(addr)
	}

	// Close any peers that were removed.
	for addr, peer := range existingPeers {
		if _, ok := newPeers[addr]; !ok {
			// TODO(ryanfowler): How do we handle an error here?
			_ = peer.Close()
		}
	}

	c.mu.Lock()
	c.hash = newHash
	c.peers = newPeers
	c.mu.Unlock()
}
