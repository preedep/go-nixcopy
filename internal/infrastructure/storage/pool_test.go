package storage

import (
	"sync"
	"testing"
	"unsafe"
)

func TestCopyBufPool_GetPut(t *testing.T) {
	buf1 := getCopyBuf()
	if len(buf1) != copyBufSize {
		t.Fatalf("getCopyBuf len = %d, want %d", len(buf1), copyBufSize)
	}
	putCopyBuf(buf1)

	buf2 := getCopyBuf()
	if len(buf2) != copyBufSize {
		t.Fatalf("second getCopyBuf len = %d, want %d", len(buf2), copyBufSize)
	}
	// After a Put/Get cycle the pool should return the same backing array.
	// This is not guaranteed under high concurrency, so we log rather than fail.
	if unsafe.SliceData(buf1) != unsafe.SliceData(buf2) {
		t.Log("pool returned a different backing array (acceptable on concurrent runs)")
	}
	putCopyBuf(buf2)
}

func TestCopyBufPool_PutSmallBuf_NotReused(t *testing.T) {
	// A resliced buffer with cap < copyBufSize must not be returned to the pool.
	small := make([]byte, 1024)
	putCopyBuf(small) // should be a no-op

	// The next get must still return a full-size buffer.
	buf := getCopyBuf()
	defer putCopyBuf(buf)
	if len(buf) != copyBufSize {
		t.Fatalf("getCopyBuf after small Put: len = %d, want %d", len(buf), copyBufSize)
	}
}

func TestCopyBufPool_ConcurrentAccess(t *testing.T) {
	const goroutines = 20
	const iters = 100
	errCh := make(chan string, goroutines)
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iters; j++ {
				buf := getCopyBuf()
				if len(buf) != copyBufSize {
					errCh <- "unexpected buffer length"
					putCopyBuf(buf)
					return
				}
				// Mutate the buffer to expose data races under -race.
				buf[0] = byte(id ^ j)
				putCopyBuf(buf)
			}
		}(i)
	}
	wg.Wait()
	close(errCh)
	if msg, ok := <-errCh; ok {
		t.Fatal(msg)
	}
}
