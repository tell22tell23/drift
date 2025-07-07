package p2p

import (
	"context"
	"log"

	libp2p "github.com/libp2p/go-libp2p"
	host "github.com/libp2p/go-libp2p/core/host"
	peer "github.com/libp2p/go-libp2p/core/peer"
)

type Host struct {
	logger *log.Logger
	host   host.Host
}

func NewHost(logger *log.Logger) (*Host, error) {
	h, err := libp2p.New()
	if err != nil {
		return nil, err
	}

	logger.Printf("Host created with ID: %s\n", h.ID())
	for _, addr := range h.Addrs() {
		logger.Printf("Listening on: %s/p2p/%s\n", addr, h.ID())
	}

	return &Host{
		logger: logger,
		host:   h,
	}, nil
}

func (h *Host) ConnectToPeer(ctx context.Context, addr string) (peer.ID, error) {
	info, err := peer.AddrInfoFromString(addr)
	if err != nil {
		return "", err
	}
	if err := h.host.Connect(ctx, *info); err != nil {
		return "", err
	}
	return info.ID, nil
}
