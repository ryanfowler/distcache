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

package grpc

import (
	"context"
	"errors"

	"github.com/ryanfowler/distcache"
	"google.golang.org/grpc"
)

type PeerCreator struct {
	DialOptions []grpc.DialOption
}

func (pc *PeerCreator) NewPeer(addr string) distcache.Peer {
	return NewClient(context.Background(), addr, pc.DialOptions...)
}

const maxRequestCount = 10

var errMaxRequestCountExceeded = errors.New("max peer request count exceeded")

type requestCountKeyType int

const requestCountKey requestCountKeyType = 0

func getRequestCount(ctx context.Context) int {
	count, _ := ctx.Value(requestCountKey).(int)
	return count
}

func withRequestCount(ctx context.Context, count int) context.Context {
	return context.WithValue(ctx, requestCountKey, count)
}
