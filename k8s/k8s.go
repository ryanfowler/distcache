package k8s

import (
	"context"
	"fmt"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	wg         sync.WaitGroup
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

func New(ctx context.Context, opts Options) (*Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	client, err := v1.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	c := &Client{
		client:     client,
		namespace:  opts.Namespace,
		onError:    opts.OnError,
		onNewPeers: opts.OnNewPeers,
		peerSetter: opts.PeerSetter,
		portName:   opts.PortName,
	}

	if err = c.start(ctx); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) Wait() {
	c.wg.Wait()
}

func (c *Client) start(ctx context.Context) error {
	i, err := c.client.Endpoints(c.namespace).Watch(ctx, metav1.ListOptions{Watch: true})
	if err != nil {
		return err
	}

	if err = c.setPeers(ctx); err != nil {
		return err
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case <-ctx.Done():
				i.Stop()
				return
			case <-i.ResultChan():
			}

			// Empty the reuslt chan before fetching peers.
			for {
				select {
				case <-i.ResultChan():
					continue
				default:
				}
				break
			}

			for {
				err := c.setPeers(ctx)
				if err == nil {
					break
				}
				if c.onError != nil {
					c.onError(err)
				}
				select {
				case <-ctx.Done():
				case <-time.After(time.Second):
					continue
				}
				break
			}
		}
	}()

	return nil
}

func (c *Client) setPeers(ctx context.Context) error {
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
