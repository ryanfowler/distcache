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
	"net"

	"github.com/ryanfowler/distcache"
	pb "github.com/ryanfowler/distcache/grpc/peerpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ (pb.PeerServer) = (*Server)(nil)

type Cache interface {
	Get(ctx context.Context, key string) ([]byte, distcache.ResultSource, error)
}

type Server struct {
	Cache Cache
	pb.UnimplementedPeerServer
}

func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	ctx = withRequestCount(ctx, int(req.GetPeerRequestCount()))
	val, res, err := s.Cache.Get(ctx, req.GetKey())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	cacheHit := res == distcache.ResultHotCache || res == distcache.ResultLocalCache
	return &pb.GetResponse{Value: val, CacheHit: cacheHit}, nil
}

func (s *Server) Listen(ctx context.Context, addr string, opt ...grpc.ServerOption) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	grpcServer := grpc.NewServer(opt...)
	pb.RegisterPeerServer(grpcServer, s)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
	}()

	return grpcServer.Serve(lis)
}
