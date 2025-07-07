package p2p

import (
	"context"
	"io"
	"log"
	"os"

	host "github.com/libp2p/go-libp2p/core/host"
	peer "github.com/libp2p/go-libp2p/core/peer"
)

type Sender struct {
	logger *log.Logger
	host   host.Host
	peerID peer.ID
}

func NewSender(logger *log.Logger, h host.Host, peerID peer.ID) *Sender {
	return &Sender{
		logger: logger,
		host:   h,
		peerID: peerID,
	}
}

func (s *Sender) SendFile(ctx context.Context, filepath string) error {
	stream, err := s.host.NewStream(ctx, s.peerID, FileTransferProtocol)
	if err != nil {
		return err
	}
	defer stream.Close()

	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	n, err := io.Copy(stream, file)
	if err != nil {
		return err
	}

	log.Printf("Sent %d bytes to %s\n", n, s.peerID)
	return nil
}
