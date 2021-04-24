package main

import (
	"context"
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
	golog.Infof("listener starting [%s] ...", desktop_laptop_connection.ListenerAddress)

	ctx, cancelFunc := context.WithCancel(context.Background())

	go listenForConnections(ctx)

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
}

func listenForConnections(ctx context.Context) {
	listener, err := net.Listen("tcp", desktop_laptop_connection.ListenerAddress)
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

			go handleNewConnection(conn)
		}
	}()

	golog.Debug("listener waiting for context done")
	<-ctx.Done()
	golog.Debug("listener got context done")
}

func handleNewConnection(conn net.Conn) {
	defer func() {
		//golog.Infof("closing connection with %s", conn.RemoteAddr())
		if err := conn.Close(); err != nil {
			golog.Errorf("failed to properly close connection %s: %s", conn.RemoteAddr(), err)
		}
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
