package main

import (
	"context"
	"fmt"
	"net"

	"github.com/kataras/golog"
)

func main() {
	golog.SetLevel(golog.DebugLevel.String())
	golog.Info("starting ...")

	ctx, cancel := context.WithCancel(context.Background())
	serverAddr, err := echoServerUDP(ctx, "127.0.0.1:")
	if err != nil {
		golog.Fatal(err)
	}
	defer cancel()

	golog.Debugf("server at [%s] listening ...", serverAddr)
	client, err := net.ListenPacket("udp", "127.0.0.1:")
	if err != nil {
		golog.Fatal(err)
	}
	defer func() { _ = client.Close() }()

	golog.Debug("sending ping ...")
	msg := []byte("ping")
	_, err = client.WriteTo(msg, serverAddr)
	if err != nil {
		golog.Fatal(err)
	}

	buf := make([]byte, 1024)
	n, addr, err := client.ReadFrom(buf)
	if err != nil {
		golog.Fatal(err)
	}

	if addr.String() != serverAddr.String() {
		golog.Fatalf("received reply from %q instead of %q", addr, serverAddr)
	} else {
		golog.Infof("received reply from correct address: %s", addr)
	}

	expectedReply := fmt.Sprintf("echo: %s", msg)
	if expectedReply != string(buf[:n]) {
		golog.Errorf("expected reply %q; actual reply %q", msg, buf[:n])
	} else {
		golog.Infof("received expected reply: %s", expectedReply)
	}
}

func echoServerUDP(ctx context.Context, addr string) (net.Addr, error) {
	packetConn, err := net.ListenPacket("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("binding to udp %s: %w", addr, err)
	}

	go func() {
		go func() {
			<-ctx.Done()
			if err := packetConn.Close(); err != nil {
				golog.Error(err)
			}
		}()

		buf := make([]byte, 1024)

		for {
			n, clientAddr, err := packetConn.ReadFrom(buf) // client to server
			if err != nil {
				return
			}

			response := fmt.Sprintf("echo: %s", buf[:n])
			_, err = packetConn.WriteTo([]byte(response), clientAddr) // server to client
			if err != nil {
				return
			}
		}
	}()

	return packetConn.LocalAddr(), nil
}
