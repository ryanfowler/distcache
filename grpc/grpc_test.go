package grpc

import (
	"bytes"
	"context"
	"errors"
	"net"
	"strings"
	"sync"
	"testing"

	"github.com/ryanfowler/distcache"
	pb "github.com/ryanfowler/distcache/grpc/peerpb/v1"
	"google.golang.org/grpc"
)

func TestGRPC(t *testing.T) {
	table := []struct {
		name            string
		getFn           func(context.Context, string) ([]byte, distcache.ResultSource, error)
		key             string
		expBytes        []byte
		expResultSource distcache.ResultSource
		expErr          bool
	}{
		{
			name: "should return successfully from cache",
			getFn: func(ctx context.Context, key string) ([]byte, distcache.ResultSource, error) {
				return []byte(key), distcache.ResultHotCache, nil
			},
			key:             "keyboard cat",
			expBytes:        []byte("keyboard cat"),
			expResultSource: distcache.ResultPeerCache,
			expErr:          false,
		},
		{
			name: "should return successfully from peer",
			getFn: func(ctx context.Context, key string) ([]byte, distcache.ResultSource, error) {
				return []byte(key), distcache.ResultLocalGet, nil
			},
			key:             "keyboard cat",
			expBytes:        []byte("keyboard cat"),
			expResultSource: distcache.ResultPeerGet,
			expErr:          false,
		},
		{
			name: "should return an error",
			getFn: func(ctx context.Context, key string) ([]byte, distcache.ResultSource, error) {
				return nil, distcache.ResultNone, errors.New("failed")
			},
			key:    "keyboard cat",
			expErr: true,
		},
	}

	ctx := context.Background()
	for i := 0; i < len(table); i++ {
		test := table[i]
		t.Run(test.name, func(t *testing.T) {
			var wg sync.WaitGroup
			defer wg.Wait()

			grpcServer := grpc.NewServer()
			defer grpcServer.GracefulStop()

			server := Server{Cache: &mockCache{getFn: test.getFn}}
			pb.RegisterPeerServiceServer(grpcServer, &server)

			lis, err := net.Listen("tcp", ":0")
			if err != nil {
				t.Fatalf("unable to create listener: %s", err.Error())
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = grpcServer.Serve(lis)
			}()

			client := NewClient(ctx, lis.Addr().String(), grpc.WithInsecure())
			defer client.Close()

			out, res, err := client.Get(ctx, test.key)
			if err != nil {
				if !test.expErr {
					t.Fatalf("unexpected error from Get: %s", err.Error())
				}
				return
			}
			if test.expErr {
				t.Fatal("unexpected success from Get")
			}
			if !bytes.Equal(test.expBytes, out) {
				t.Fatalf("unexpected bytes returned: %s", out)
			}
			if test.expResultSource != res {
				t.Fatalf("unexpected result source: %s", res.String())
			}
		})
	}
}

func TestGRPCLoop(t *testing.T) {
	addr := getFreeAddr(t)

	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := NewClient(ctx, addr, grpc.WithInsecure())
	defer client.Close()

	server := Server{Cache: &mockCache{
		getFn: func(ctx context.Context, key string) ([]byte, distcache.ResultSource, error) {
			return client.Get(ctx, key)
		},
	}}

	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = server.Listen(ctx, addr)
	}()

	_, _, err := client.Get(ctx, "keyboard cat")
	if err == nil || !strings.Contains(err.Error(), errMaxRequestCountExceeded.Error()) {
		t.Fatalf("unexpected error returned: %v", err)
	}
}

func getFreeAddr(t *testing.T) string {
	t.Helper()
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("unable to listen: %s", err.Error())
	}
	defer lis.Close()
	return lis.Addr().String()
}

type mockCache struct {
	getFn func(context.Context, string) ([]byte, distcache.ResultSource, error)
}

func (c *mockCache) Get(ctx context.Context, key string) ([]byte, distcache.ResultSource, error) {
	return c.getFn(ctx, key)
}
