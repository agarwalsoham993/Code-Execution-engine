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
	if cb.Len() >= cb.cap {
		return len(p), nil 								// Already full, pretend we wrote it
	}
	
	remaining := cb.cap - cb.Len()
	if len(p) > remaining {								// write only what fits
		cb.Buffer.Write(p[:remaining])
		return len(p), nil 
	}
	
	return cb.Buffer.Write(p)
}