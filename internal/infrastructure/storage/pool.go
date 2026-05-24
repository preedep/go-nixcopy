package storage

import "sync"

// copyBufPool pools 256 KiB buffers for io.CopyBuffer calls across all storage backends.
// Using a pool avoids a per-transfer heap allocation and reduces GC pressure during batch transfers.
const copyBufSize = 256 * 1024

var copyBufPool = sync.Pool{
	New: func() any {
		buf := make([]byte, copyBufSize)
		return &buf
	},
}

func getCopyBuf() []byte {
	return *copyBufPool.Get().(*[]byte)
}

func putCopyBuf(buf []byte) {
	// Only return full-size buffers to avoid pool pollution with resliced slices.
	if cap(buf) == copyBufSize {
		buf = buf[:copyBufSize]
		copyBufPool.Put(&buf)
	}
}
