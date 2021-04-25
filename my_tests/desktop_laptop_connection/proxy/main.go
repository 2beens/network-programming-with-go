package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"

	"github.com/kataras/golog"
)

func main() {
	golog.SetLevel("debug")

	listenAddr := flag.String("addr", "", "listen address, e.g. 192.168.178.1")
	listenPort := flag.String("port", "9999", "listen port, e.g. 9999")
	flag.Parse()
	endpoint := fmt.Sprintf("%s:%s", *listenAddr, *listenPort)

	golog.Infof("proxy listener starting [%s] ...", endpoint)

	ctx, cancelFunc := context.WithCancel(context.Background())

	go listenForConnections(ctx, endpoint)

	chOsInterrupt := make(chan os.Signal, 1)
	signal.Notify(chOsInterrupt, os.Interrupt)
	<-chOsInterrupt

	golog.Warn("caught interrupt signal, will quit")
	cancelFunc()
}

func listenForConnections(ctx context.Context, endpoint string) {
	listener, err := net.Listen("tcp", endpoint)
	if err != nil {
		golog.Fatal(err)
	}
	golog.Debug("proxy listener created")

	var server net.Conn = nil
	var client net.Conn = nil // TODO: make more connections possible later

	defer func() {
		if server != nil {
			server.Close()
		}
		if client != nil {
			client.Close()
		}
		golog.Warn("proxy, bye bye ...")
	}()

	go func() {
		for {
			if err := ctx.Err(); err == context.Canceled {
				break
			} else if err != nil {
				golog.Warnf("context canceled: %s", err)
				break
			}

			golog.Debug("listening for connections ...")
			conn, err := listener.Accept()
			if err != nil {
				golog.Warnf("failed to accept connection: %s", err)
				continue
			}
			golog.Debugf("new connection received ...")

			buf := make([]byte, 32)
			n, err := conn.Read(buf)
			if err != nil {
				if err != io.EOF {
					golog.Error(err)
				}
				break
			}

			connType := string(buf[:n])
			switch connType {
			case "client":
				golog.Debugf("got new client connection: %s", conn.RemoteAddr())
				client = conn
			case "server":
				golog.Debugf("got new server connection: %s", conn.RemoteAddr())
				server = conn
			default:
				golog.Errorf("unrecognized conn type received: %q", connType)
			}
		}
	}()

	go func() {
		for {
			if server == nil || client == nil {
				continue
			}

			err = proxy(client, server)
			if err != nil && err != io.EOF {
				golog.Errorf("failed to proxy request: %s", err)
			}
		}
	}()

	golog.Info("proxy listener waiting for context done")
	<-ctx.Done()

	golog.Warn("proxy listener got context done, closing connections ...")
	if server != nil {
		if err := server.Close(); err != nil {
			golog.Error(err)
		}
	}
	if client != nil {
		if err := client.Close(); err != nil {
			golog.Error(err)
		}
	}
}

func proxy(from io.Reader, to io.Writer) error {
	fromWriter, fromIsWriter := from.(io.Writer)
	toReader, toIsReader := to.(io.Reader)

	if toIsReader && fromIsWriter {
		// send replies since "from" and "to" implement the necessary interfaces
		go func() { _, _ = io.Copy(fromWriter, toReader) }()
	}

	_, err := io.Copy(to, from)
	return err
}
