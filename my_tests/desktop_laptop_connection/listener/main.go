package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/2beens/network-programming-with-go/my_tests/desktop_laptop_connection"
	"github.com/gosuri/uiprogress"
	"github.com/kataras/golog"
)

func main() {
	golog.SetLevel("debug")

	listenAddr := flag.String("addr", "", "listen address, e.g. 192.168.178.1")
	listenPort := flag.String("port", "9999", "listen port, e.g. 9999")
	proxyAddr := flag.String("proxy", "", "route via proxy (empty for no proxy), e.g. 127.0.0.1:9998")
	flag.Parse()
	endpoint := fmt.Sprintf("%s:%s", *listenAddr, *listenPort)

	golog.Infof("listener starting [%s] ...", endpoint)

	ctx, cancelFunc := context.WithCancel(context.Background())

	go listenForConnections(ctx, endpoint, *proxyAddr)

	localIp, err := desktop_laptop_connection.GetLocalIp()
	if err != nil {
		golog.Errorf("failed to get local ip: %s", err)
	} else {
		golog.Warnf("local ip: %s", localIp)
	}

	chOsInterrupt := make(chan os.Signal, 1)
	signal.Notify(chOsInterrupt, os.Interrupt)
	<-chOsInterrupt

	golog.Warn("caught interrupt signal, will quit")
	cancelFunc()

	// TODO: use WaitGroup to first wait for all open connections to close, and then quit
}

func listenForConnections(ctx context.Context, endpoint, proxyAddr string) {
	listener, err := net.Listen("tcp", endpoint)
	if err != nil {
		golog.Fatal(err)
	}
	golog.Debug("listener created")

	uiprogress.Start()

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

			go handleNewConnection(ctx, conn)
		}
	}()

	go func() {
		if proxyAddr == "" {
			return
		}
		golog.Debugf("using proxy: %s", proxyAddr)

		conn, err := net.Dial("tcp", proxyAddr)
		if err != nil {
			golog.Fatal(err)
		}
		golog.Infof("proxy connection established, sending server info...")

		_, err = conn.Write([]byte("server"))
		if err != nil {
			golog.Error(err)
			conn.Close()
			return
		}

		go handleNewConnection(ctx, conn)
	}()

	golog.Debug("listener waiting for context done")
	<-ctx.Done()
	golog.Debug("listener got context done")
}

func handleNewConnection(ctx context.Context, conn net.Conn) {
	connClosed := false
	defer func() {
		if connClosed {
			return
		}
		//golog.Infof("closing connection with %s", conn.RemoteAddr())
		if err := conn.Close(); err != nil {
			golog.Errorf("failed to properly close connection %s: %s", conn.RemoteAddr(), err)
		}
	}()

	go func() {
		<-ctx.Done()
		if err := conn.Close(); err != nil {
			golog.Errorf("failed to properly close connection %s: %s", conn.RemoteAddr(), err)
		}
		connClosed = true
	}()

	golog.Debugf("got new connection: %s", conn.RemoteAddr())

	bar := uiprogress.AddBar(100)
	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return fmt.Sprintf("%s", conn.RemoteAddr().String())
	})
	if err := bar.Set(50); err != nil {
		golog.Error(err)
	}

	closedReasonMessage := "closed!"
	buf := make([]byte, 1)
	for {
		// every time we receive a new input from the client, we move the read/write deadline
		if err := conn.SetDeadline(time.Now().Add(time.Minute)); err != nil {
			closedReasonMessage = fmt.Sprintf("failed to set read/write deadline for %s: %s", conn.RemoteAddr(), err)
			break
		}

		_, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				closedReasonMessage = fmt.Sprintf("failed to read data from connection %s: %s", conn.RemoteAddr(), err)
			}
			break
		}
		//golog.Debugf("read %d bytes: %q", n, string(buf))

		controlVal := string(buf)
		if controlVal == desktop_laptop_connection.ControlIncrease {
			bar.Incr()
			continue
		}
		if controlVal == desktop_laptop_connection.ControlDecrease {
			if err := bar.Set(bar.Current() - 1); err != nil {
				golog.Error(err)
			}
			continue
		}

		closedReasonMessage = fmt.Sprintf("unknown control value: %+v", controlVal)
		break
	}

	bar.AppendFunc(func(b *uiprogress.Bar) string {
		return closedReasonMessage
	})
}
