package unix_socket

type UnixSocketPayload []struct {
}

func (u *UnixSocketPayload) Encode() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (u *UnixSocketPayload) Decode(data []byte) error {
	//TODO implement me
	panic("implement me")
}

func (u *UnixSocketPayload) Length() int {
	//TODO implement me
	panic("implement me")
}
