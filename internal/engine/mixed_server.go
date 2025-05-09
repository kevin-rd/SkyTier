package engine

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"kevin-rd/my-tier/internal/seed"
	"kevin-rd/my-tier/pkg/packet"
	"log"
	"net"
)

// StartMixedServer start mixed server
func (e *Engine) StartMixedServer() error {
	mixedAddr := fmt.Sprintf(":%d", e.config.MixedPort)
	log.Println("start mixed server fixed:", mixedAddr)
	listener, err := net.Listen("tcp", mixedAddr)
	if err != nil {
		return fmt.Errorf("listen error: %w", err)
	}

	e.mux = &ServeMux{handlers: make(map[byte]HandlerFunc)}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println("accept error:", err)
				continue
			}
			go e.handleConn(conn)
		}
	}()
	return nil
}

func (e *Engine) handleConn(conn net.Conn) {
	defer func(conn net.Conn) {
		if err := conn.Close(); err != nil {
			log.Println("close conn error on defer:", err)
		}
	}(conn)

	buf := make([]byte, 1500)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Println("client closed connection:", err)
			} else {
				log.Println("mixed server read error:", err)
			}
			return
		}

		var pkt *packet.Packet
		pkt, err = packet.Decode(buf[:n])
		if err != nil {
			log.Println("decode error:", err)
			continue
		}
		e.mux.serve(conn, pkt)
	}
}

// --- handlers

// HandleGetPeers handle the get peers request
func (e *Engine) HandleGetPeers(w ResponseWriter, r *Request) {
	// authentication
	// todo

	networkName := "default"
	peers, err := e.seed.GetPeers(networkName)
	if err != nil {
		return
	}

	if _, err = w.WritePayload(packet.TypeAuxPeersReply, &seed.PeersReplyPayload{
		Peers: peers,
	}); err != nil {
		log.Printf("write payload error: %v", err)
		return
	}
}

func (e *Engine) HandlePing(w ResponseWriter, r *Request) {
	if _, err := w.WritePayload(packet.TypePong, "Pong"); err != nil {
		log.Printf("write pong error: %v", err)
		return
	}
}

// ----- ServeMux -----

// Request
type Request struct {
	Conn    net.Conn       // 原始连接
	Packet  *packet.Packet // 原始 Packet
	Payload []byte         // 提取 Payload
}

type ResponseWriter interface {
	Write(pkt []byte) (int, error)
	WriteP(pkt *packet.Packet) (int, error)

	WritePayload(typ byte, payload any) (int, error)
}

type response struct {
	conn net.Conn
}

func (w *response) Write(pkt []byte) (int, error) {
	return w.conn.Write(pkt)
}

func (w *response) WriteP(pkt *packet.Packet) (int, error) {
	bts, err := pkt.Encode()
	if err != nil {
		return 0, err
	}
	return w.conn.Write(bts)
}

func (w *response) WritePayload(typ byte, payload any) (int, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(payload); err != nil {
		return 0, err
	}

	return w.WriteP(&packet.Packet{
		Version: packet.ProtocolVersion,
		Type:    typ,
		Payload: buf.Bytes(),
	})
}

type HandlerFunc func(w ResponseWriter, r *Request)

// ServeMux
type ServeMux struct {
	handlers map[byte]HandlerFunc
}

// RegisterHandle register handle func
func (m *ServeMux) RegisterHandle(packetType byte, h HandlerFunc) {
	m.handlers[packetType] = h
}

func (m *ServeMux) serve(conn net.Conn, pkt *packet.Packet) {
	w := &response{conn: conn}
	r := &Request{Conn: conn, Packet: pkt, Payload: pkt.Payload}

	if h, ok := m.handlers[pkt.Type]; ok {
		h(w, r)
	} else {
		log.Printf("unknown packet type: %d\n", pkt.Type)
	}
}
