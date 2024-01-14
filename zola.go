package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
)

const buffer32K = 32 * 1024

var wg sync.WaitGroup

var (
	buffer32KPool = newBufferPoolWithSize(buffer32K)
)

// BufioReaderPool is a bufio reader that uses sync.Pool.
type BufioReaderPool struct {
	pool sync.Pool
}

// newBufioReaderPoolWithSize is unexported because new pools should be
// added here to be shared where required.
func newBufioReaderPoolWithSize(size int) *BufioReaderPool {
	return &BufioReaderPool{
		pool: sync.Pool{
			New: func() interface{} { return bufio.NewReaderSize(nil, size) },
		},
	}
}

// Get returns a bufio.Reader which reads from r. The buffer size is that of the pool.
func (bufPool *BufioReaderPool) Get(r io.Reader) *bufio.Reader {
	buf := bufPool.pool.Get().(*bufio.Reader)
	buf.Reset(r)
	return buf
}

// Put puts the bufio.Reader back into the pool.
func (bufPool *BufioReaderPool) Put(b *bufio.Reader) {
	b.Reset(nil)
	bufPool.pool.Put(b)
}

type bufferPool struct {
	pool sync.Pool
}

func newBufferPoolWithSize(size int) *bufferPool {
	return &bufferPool{
		pool: sync.Pool{
			New: func() interface{} { s := make([]byte, size); return &s },
		},
	}
}

func (bp *bufferPool) Get() *[]byte {
	return bp.pool.Get().(*[]byte)
}

func (bp *bufferPool) Put(b *[]byte) {
	bp.pool.Put(b)
}

func Copy(dst io.Writer, src io.Reader) (written int64, err error) {
	currTime := time.Now()
	buf := buffer32KPool.Get()
	written, err = io.CopyBuffer(dst, src, *buf)
	fmt.Println("lola trying in Copy over: ", written, err, time.Since(currTime).Milliseconds())
	buffer32KPool.Put(buf)
	return
}

func main() {
	// create a wait group

	copyFunc := func(w io.Writer, r io.ReadCloser) {
		wg.Add(1)
		go func() {
			if _, err := Copy(w, r); err != nil {
				fmt.Println("Lola: error", err)
			}
			fmt.Println("End")
			r.Close()
			wg.Done()
		}()
	}

	cmd := exec.Command("bash", "-c", "./pola.sh")
	stdOutPipe, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("Lola: stdoutErr", err)
	}

	// stderrPipe, err := cmd.StderrPipe()
	// if err != nil {
	// 	fmt.Println("Lola: stderrpipeErr", err)
	// }

	fmt.Println("Executing")

	// copyFunc(os.Stderr, stderrPipe)
	copyFunc(os.Stdout, stdOutPipe)

	cmd.Run()

	cmd.Wait()

	wg.Wait()
	fmt.Println("Lola: wait done")
}
