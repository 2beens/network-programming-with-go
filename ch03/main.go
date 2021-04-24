package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	fmt.Println("starting ...")
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Printf("## failed to create a TCP listener: %s\n", err)
		return
	}

	defer func() {
		if err := listener.Close(); err != nil {
			fmt.Printf("## failed to close TCP listener: %s\n", err)
			return
		}
		fmt.Println("listener closed")
	}()

	fmt.Printf("bound to %q\n", listener.Addr())

	conn, err := listener.Accept()
	if err != nil {
		fmt.Printf("## failed to accept a connection: %s\n", err)
		return
	}

	fmt.Println("connection accepted ...")

	if err := conn.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		fmt.Printf("## failed to set read deadline: %s\n", err)
		return
	}

	var readBytes []byte
	bytesRead, err := conn.Read(readBytes)
	if err != nil {
		fmt.Printf("## failed to read from connection: %s\n", err)
		return
	}

	fmt.Printf("connection read [%d] bytes of: %s\n", bytesRead, readBytes)

	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Printf("## failed to close connection: %s\n", err)
			return
		}
		fmt.Println("connection closed")
	}()
}
