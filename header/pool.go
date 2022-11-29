package header

import "sync"

var (
	ReqeustPool  sync.Pool
	ResponsePool sync.Pool
)

func init() {
	ReqeustPool = sync.Pool{New: func() any {
		return &RequestHeader{}
	}}
	ResponsePool = sync.Pool{New: func() any {
		return &ResponseHeader{}
	}}
}
