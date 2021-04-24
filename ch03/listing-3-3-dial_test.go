package main

import (
	"io"
	"net"
	"testing"
)

func TestDial(t *testing.T) {
	// Create a listener on a random port.
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("listener created at: %s", listener.Addr())

	done := make(chan struct{})
	go func() {
		defer func() { done <- struct{}{} }()

		connCounter := 0
		for {
			connCounter++
			t.Logf("creating connection %d ...", connCounter)

			conn, err := listener.Accept()
			if err != nil {
				t.Log(err)
				return
			}

			t.Logf("connection %d created", connCounter)

			go func(c net.Conn, connCounter int) {
				defer func() {
					c.Close()
					done <- struct{}{}
				}()

				buf := make([]byte, 1024)
				for {
					n, err := c.Read(buf)
					if err != nil {
						if err != io.EOF {
							t.Error(err)
						}
						t.Logf("connection %d received EOF", connCounter)
						return
					}
					t.Logf("received: %q", buf[:n])
				}
			}(conn, connCounter)
		}
	}()

	clientConn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	//clientConn.Write([]byte("ping-st"))

	t.Log("-> closing clientConn ...")
	clientConn.Close()
	<-done
	t.Log("-> closing listener ...")
	listener.Close()
	<-done
	t.Log("-> all done")
}
