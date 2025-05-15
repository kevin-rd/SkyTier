package unix_socket

import (
	"kevin-rd/my-tier/internal/ipc"
	"kevin-rd/my-tier/internal/peer"
	"kevin-rd/my-tier/pkg/ipc/message"
	"log"
)

func (_ *UnixSocket) HandleGetPeers(fGet func(network string) []*peer.Peer) ipc.Handler {
	return func(writer message.Writer, r *message.Message) {
		body, err := message.DecodePayload[message.PeersReq](r)
		if err != nil {
			return
		}
		log.Printf("[unixsocket] get peers request: %v", body)

		msg, err := message.New(message.KindPeers, &message.PeersResp{
			Peers: fGet(""),
		})
		if err != nil {
			log.Printf("[unixsocket] new message error: %v", err)
			return
		}

		if err := writer.Write(msg); err != nil {
			log.Printf("[unixsocket] write error: %v", err)
			return
		}
	}
}
