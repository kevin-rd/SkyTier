package unix_socket

import (
	"encoding/json"
	"errors"
	"kevin-rd/my-tier/pkg/ipc/message"
	"log"
	"net"
)

const (
	UNIX_SOCKET_PATH = "/tmp/skytier-core.sock"
)

func Get(msg *message.Message) (*message.Message, error) {
	conn, err := net.Dial("unix", UNIX_SOCKET_PATH)
	if err != nil {
		log.Printf("dial error: %v", err)
		return nil, errors.Join(err, errors.New("dial error"))
	}
	defer conn.Close()

	writer := message.NewWriter(conn)
	if err = writer.Write(msg); err != nil {
		return nil, errors.Join(errors.New("write error"), err)
	}

	decoder := json.NewDecoder(conn)
	var resp message.Message
	if err = decoder.Decode(&resp); err != nil {
		return nil, errors.Join(errors.New("decode error"), err)
	}

	// todo: 二次unmarshal

	if resp.Kind != message.KindPeers {
		return nil, errors.New("invalid response kind")
	}
	return &resp, nil
}
