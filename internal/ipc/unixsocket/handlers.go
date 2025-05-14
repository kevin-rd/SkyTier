package unix_socket

import (
	"kevin-rd/my-tier/internal/ipc"
	"kevin-rd/my-tier/internal/peer"
	"kevin-rd/my-tier/pkg/ipc/message"
	"log"
)

func (_ *UnixSocket) HandleGetPeers(fGet func(network string) []*peer.Peer) ipc.Handler {
	return func(writer message.Writer, _ *message.Message) {
		msg := &message.Message{
			Kind: message.KindPeers,
			Payload: &message.PeersResp{
				Peers: fGet(""),
			},
		}

		if err := writer.Write(msg); err != nil {
			log.Printf("[unixsocket] write error: %v", err)
			return
		}
	}
}
