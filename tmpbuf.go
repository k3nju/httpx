package httpx

type TmpBuf struct {
	size uint64 // alloc size
	buf  []byte // current buffer
	wi   uint64 // write index
}

func NewTmpBuf(size uint64) *TmpBuf {
	buf := make([]byte, size)
	return &TmpBuf{
		size: size,
		buf:  buf,
		wi:   0,
	}
}

func (t *TmpBuf) Prepare(size uint64) ([]byte, []byte) {
	w := t.buf[t.wi:]
	if len(w) >= int(size) {
		// enough space to cpy "size" bytes to t.buf[t.wi:]
		return w[:size], nil
	}

	// no space to cpoy size bytes
	// detach current buffer and allocate new buffer
	var detached []byte
	if t.wi > 0 {
		detached = t.buf[:t.wi]
	}
	t.buf = make([]byte, t.size)
	t.wi = 0

	if size > t.size {
		// when reallocated but more space required, return nil
		return nil, detached
	}

	return t.buf[:size], detached
}

func (t *TmpBuf) Consume(size uint64) {
	t.wi += size
}

func (t *TmpBuf) Detach() []byte {
	detached := t.buf[:t.wi]
	t.buf = t.buf[t.wi:]
	t.wi = 0

	return detached
}
