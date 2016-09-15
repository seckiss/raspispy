package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"runtime"
	"time"

	"github.com/seckiss/ringbuf"
)

func main() {
	r := bufio.NewReader(os.Stdin)
	var ring = ringbuf.NewBuffer(5 * 1024 * 1024)
	go ringWrite(ring, r)
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
		conn.SetDeadline(time.Now().Add(30 * time.Second))
		go func(c net.Conn) {
			reader := ring.NewReader()
			written, err := io.Copy(c, reader)
			if err != nil {
				fmt.Printf("io.Copy ended with error: %+v\n", err)
			}
			fmt.Printf("Written %d bytes\n", written)

			/*			var b = make([]byte, 0, 4*1024)
						for {
							n, err := reader.Read(b[:cap(b)])
							if err != nil {
								break
							}
							if n == 0 {
								continue
							}
							var b2 = make([]byte, 0, n)
							conn.Write(b2)

						}
			*/
			c.Close()

		}(conn)

	}

	//time.Sleep(20 * time.Second)
	//ioutil.WriteFile("out.rasp", ring.Bytes(), 0644)
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
