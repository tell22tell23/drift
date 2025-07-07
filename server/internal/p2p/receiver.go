package p2p

import (
	"io"
	"log"
	"os"

	"github.com/libp2p/go-libp2p/core/network"
)

func FileReceiverHandler(s network.Stream, log *log.Logger) {
	log.Printf("Incoming file transfer request from %s\n", s.Conn().RemotePeer())

	defer s.Close()
	out, err := os.Create("received.txt")
	if err != nil {
		log.Printf("Error creating file: %v\n", err)
		return
	}
	defer out.Close()

	n, err := io.Copy(out, s)
	if err != nil {
		log.Printf("Error receiving file: %v\n", err)
		return
	}
	log.Printf("Received %d bytes from %s\n", n, s.Conn().RemotePeer())
}
