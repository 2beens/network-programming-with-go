package main

import (
	"crypto/rand"
	"io"
	"log"
	"net"
	"os"
)

func main() {
	log.SetOutput(os.Stdout)

	payload := make([]byte, 1<<24) // 16 MB
	_, err := rand.Read(payload)   // generate a random payload
	if err != nil {
		log.Fatal(err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close()

		_, err = conn.Write(payload)
		if err != nil {
			log.Println(err)
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	buf := make([]byte, 1<<19) // 512 KB
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Println(err)
			}
			break
		}
		log.Printf("read %d bytes", n) // buf[:n] is the data read from conn
	}
}
