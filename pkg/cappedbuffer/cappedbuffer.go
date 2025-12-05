package cappedbuffer

import "bytes"

type CappedBuffer struct {
	*bytes.Buffer
	cap int
}

func New(buf []byte, cap int) *CappedBuffer {
	return &CappedBuffer{Buffer: bytes.NewBuffer(buf), cap: cap}
}

func (cb *CappedBuffer) Write(p []byte) (n int, err error) {
	if cb.Len()+len(p) > cb.cap { return 0, nil }
	return cb.Buffer.Write(p)
}
