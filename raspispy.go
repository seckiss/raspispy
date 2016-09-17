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
	"strings"
	"time"

	"github.com/seckiss/ringbuf"
)

func main() {
	r := bufio.NewReader(os.Stdin)
	var ring = ringbuf.NewBuffer(100 * 1024 * 1024)
	go ringWrite(ring, r)
	go streamServer(ring)
	go controlServer(ring)
	select {}
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
			line = strings.TrimSpace(line)
			if line == "dump" {
				var dumpname = "raspispydump_" + time.Now().Format("060102_150405") + ".h264"
				err = ioutil.WriteFile(dumpname, ring.Bytes(), 0644)
				var response = "scp pi:/home/pi/seckiss/raspispy/" + dumpname + " . ; mpv " + dumpname + "\n"
				if err != nil {
					response = fmt.Sprintf("ioutil.WriteFile ended with error: %+v\n", err)
					fmt.Println(response)
				}
				c.Write([]byte(response))
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
