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
)

var _ (pb.PeerServer) = (*Server)(nil)

type Server struct {
	Cache *distcache.Cache
}

func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	val, res, err := s.Cache.Get(ctx, req.Key)
	if err != nil {
		// TODO(ryanfowler): Use an appropriate grpc error here.
		return nil, err
	}
	cacheHit := res == distcache.ResultHotCache || res == distcache.ResultLocalCache
	return &pb.GetResponse{Value: val, CacheHit: cacheHit}, nil
}
