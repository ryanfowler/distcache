package k8s

import (
	"context"
	"errors"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type Client struct {
	client     *v1.CoreV1Client
	namespace  string
	onError    func(error)
	onNewPeers func(...string)
	peerSetter PeerSetter
	portName   string
}

type PeerSetter interface {
	SetPeers(peers ...string)
}

type Options struct {
	Namespace  string
	PortName   string
	PeerSetter PeerSetter

	OnError    func(err error)
	OnNewPeers func(peers ...string)
}

func New(opts Options) (*Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	client, err := v1.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Client{
		client:     client,
		namespace:  opts.Namespace,
		onError:    opts.OnError,
		onNewPeers: opts.OnNewPeers,
		peerSetter: opts.PeerSetter,
		portName:   opts.PortName,
	}, nil
}

func (c *Client) RefreshPeers(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	es, err := c.client.Endpoints(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	var peers []string
	for _, item := range es.Items {
		for _, subset := range item.Subsets {
			for _, addr := range subset.Addresses {
				for _, port := range subset.Ports {
					if port.Name != c.portName {
						continue
					}
					peers = append(peers, fmt.Sprintf("%s:%d", addr.IP, port.Port))
					break
				}
			}
		}
	}

	if c.onNewPeers != nil {
		c.onNewPeers(peers...)
	}

	c.peerSetter.SetPeers(peers...)
	return nil
}

func (c *Client) Watch(ctx context.Context) error {
	for {
		err := c.watch(ctx)
		if err != nil && c.onError != nil {
			c.onError(err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
}

func (c *Client) watch(ctx context.Context) error {
	i, err := c.client.Endpoints(c.namespace).Watch(ctx, metav1.ListOptions{Watch: true})
	if err != nil {
		return err
	}
	for {
		if err = c.waitToRefresh(ctx, i); err != nil {
			return err
		}
	}
}

func (c *Client) waitToRefresh(ctx context.Context, i watch.Interface) error {
	timer := time.NewTimer(10 * time.Minute)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case _, ok := <-i.ResultChan():
		if !ok {
			return errors.New("watch channel closed")
		}
	case <-timer.C:
	}

	for {
		err := c.RefreshPeers(ctx)
		if err == nil {
			return nil
		}
		if c.onError != nil {
			c.onError(err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
}
