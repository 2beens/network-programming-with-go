package main

import (
	"bufio"
	"log"
	"net"
	"os"
)

func main() {
	log.SetOutput(os.Stdout)
	log.Println("starting ...")

	listener, err := net.Listen("tcp", "127.0.0.1:11999")
	if err != nil {
		log.Fatalln(err)
	}

	go func() {
		log.Printf("listening for incoming connections on [%s] ...", listener.Addr())

		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close()

		log.Printf("connection [%s] received, sending response", conn.RemoteAddr())

		scanner := bufio.NewScanner(conn)
		scanner.Scan()
		log.Printf("received: %s", scanner.Text())

		_, err = conn.Write([]byte("sta ima?"))
		if err != nil {
			log.Println(err)
		}
	}()

	select {}
}
