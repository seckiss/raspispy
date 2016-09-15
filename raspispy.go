package main

import (
	"bufio"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/seckiss/ringbuf"
)

func main() {
	r := bufio.NewReader(os.Stdin)
	var ring = ringbuf.NewBuffer(5 * 1024 * 1024)
	go ringWrite(ring, r)
	time.Sleep(20 * time.Second)
	ioutil.WriteFile("out.rasp", ring.Bytes(), 0644)
}

func ringWrite(ring *ringbuf.Buffer, r io.Reader) {
	var buf = make([]byte, 0, 4*1024)
	var nBytes, nChunks = int64(0), int64(0)
	for {
		n, err := r.Read(buf[:cap(buf)])
		buf = buf[:n]
		if n == 0 {
			if err == nil {
				runtime.Gosched()
				continue
			}
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		nChunks++
		nBytes += int64(len(buf))
		// process buf
		ring.Write(buf[:n])
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
		runtime.Gosched()
	}
	log.Println("Bytes:", nBytes, "Chunks:", nChunks)
}
