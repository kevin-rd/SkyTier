package mixed_server

import "sync"

var (
	bufPool = sync.Pool{
		New: func() any {
			return make([]byte, 1024)
		},
	}
)
