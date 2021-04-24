package main

import (
	"context"
	"io"
	"net"
	"os"
	"os/signal"
	"time"

	"fmt"

	"github.com/2beens/network-programming-with-go/my_tests/desktop_laptop_connection"
	"github.com/gosuri/uiprogress"
	"github.com/kataras/golog"
)

func main() {
	golog.SetLevel("debug")
	golog.Infof("listener starting [%s] ...", desktop_laptop_connection.ListenerAddress)

	ctx, cancelFunc := context.WithCancel(context.Background())

	go listenForConnections(ctx)

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
		golog.Infof("closing connection with %s", conn.RemoteAddr())
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

	//var typ int8
	//err := binary.Read(conn, binary.BigEndian, &typ) // 1-byte type
	//if err != nil {
	//	return 0, err
	//}

	buf := make([]byte, 1)
	for {
		// every time we receive a new input from the client, we move the read/write deadline
		if err := conn.SetDeadline(time.Now().Add(time.Minute)); err != nil {
			golog.Errorf("failed to set read/write deadline for %s: %s", conn.RemoteAddr(), err)
			break
		}

		_, err := conn.Read(buf)
		if err != nil && err != io.EOF {
			golog.Errorf("failed to read data from connection %s: %s", conn.RemoteAddr(), err)
			break
		}

		controlVal := string(buf)
		//golog.Debugf("read %d bytes: %s", n, controlVal)
		switch controlVal {
		case desktop_laptop_connection.ControlIncrease:
			bar.Incr()
		case desktop_laptop_connection.ControlDecrease:
			if err := bar.Set(bar.Current() - 1); err != nil {
				golog.Error(err)
			}
		default:
			golog.Errorf("unknown control value: %+v", controlVal)
		}
	}
}

func progressBarExample() {
	uiprogress.Start()            // start rendering
	bar := uiprogress.AddBar(100) // Add a new bar
	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return "app: "
	})
	for bar.Incr() {
		time.Sleep(time.Millisecond * 20)
		if bar.Current() == 50 {
			//if err := bar.Set(10); err != nil {
			//	golog.Fatal(err)
			//}
			bar.AppendFunc(func(b *uiprogress.Bar) string {
				return "done"
			})
			break
		}
	}
}
