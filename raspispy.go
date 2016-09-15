package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"runtime"

	"github.com/seckiss/ringbuf"
)

func main() {
	r := bufio.NewReader(os.Stdin)
	var ring = ringbuf.NewBuffer(30 * 1024 * 1024)
	go ringWrite(ring, r)
	go streamServer(ring)
	go controlServer(ring)
}

func streamServer(ring *ringbuf.Buffer) {
	l, err := net.Listen("tcp", ":2222")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go func(c net.Conn) {
			reader := ring.NewReaderOffset(1024 * 1024)
			written, err := io.Copy(c, reader)
			if err != nil {
				fmt.Printf("io.Copy ended with error: %+v\n", err)
			}
			fmt.Printf("Written %d bytes\n", written)
			c.Close()
		}(conn)
	}
}

func controlServer(ring *ringbuf.Buffer) {
	l, err := net.Listen("tcp", ":2223")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go func(c net.Conn) {
			reader := bufio.NewReader(c)
			line, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("ReadString ended with error: %+v\n", err)
			}
			if line == "dump" {
				err = ioutil.WriteFile("dump.h264", ring.Bytes(), 0644)
				if err != nil {
					fmt.Printf("ioutil.WriteFile ended with error: %+v\n", err)
				}

			}
			c.Close()

		}(conn)
	}
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
