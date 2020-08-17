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
	"hash/crc32"
	"sort"
	"strconv"
)

type peerHash struct {
	hash   func([]byte) int
	hashes []int
	peers  map[int]string
}

func newPeerHash(peers ...string) *peerHash {
	const vNodes = 32
	h := &peerHash{
		hash:   func(b []byte) int { return int(crc32.ChecksumIEEE(b)) },
		hashes: make([]int, 0, vNodes*len(peers)),
		peers:  make(map[int]string),
	}
	for _, peer := range peers {
		for i := 0; i < vNodes; i++ {
			hash := h.hash([]byte(strconv.Itoa(i) + "_" + peer))
			h.hashes = append(h.hashes, hash)
			h.peers[hash] = peer
		}
	}
	sort.Ints(h.hashes)
	return h
}

func (h *peerHash) GetPeer(key []byte) string {
	if len(h.hashes) == 0 {
		return ""
	}
	idx := sort.SearchInts(h.hashes, h.hash(key))
	if idx >= len(h.hashes) {
		idx = 0
	}
	return h.peers[h.hashes[idx]]
}
