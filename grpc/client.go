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

	"github.com/ryanfowler/distcache"
	pb "github.com/ryanfowler/distcache/grpc/peerpb"
	"google.golang.org/grpc"
)

var _ distcache.Peer = (*Client)(nil)

type Client struct {
	err     error
	address string
	client  pb.PeerClient
	conn    *grpc.ClientConn
}

func NewClient(ctx context.Context, addr string, opts ...grpc.DialOption) *Client {
	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return &Client{err: err}
	}
	return &Client{
		address: addr,
		client:  pb.NewPeerClient(conn),
		conn:    conn,
	}
}

func (c *Client) Get(ctx context.Context, key string) ([]byte, distcache.ResultSource, error) {
	if c.err != nil {
		return nil, distcache.ResultNone, c.err
	}

	count := getRequestCount(ctx)
	count++
	if count > maxRequestCount {
		return nil, distcache.ResultNone, errMaxRequestCountExceeded
	}

	res, err := c.client.Get(ctx, &pb.GetRequest{
		Key:              key,
		PeerRequestCount: int32(count),
	})
	if err != nil {
		return nil, distcache.ResultNone, err
	}
	resSrc := distcache.ResultPeerGet
	if res.CacheHit {
		resSrc = distcache.ResultPeerCache
	}
	return res.Value, resSrc, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) String() string {
	return c.address
}
