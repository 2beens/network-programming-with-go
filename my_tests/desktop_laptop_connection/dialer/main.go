package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/2beens/network-programming-with-go/my_tests/desktop_laptop_connection"
	"github.com/eiannone/keyboard"
	"github.com/kataras/golog"
)

func main() {
	golog.SetLevel("debug")

	listenAddr := flag.String("addr", "", "listen address, e.g. 192.168.178.1")
	listenPort := flag.String("port", "9999", "listen port, e.g. 9999")
	proxyAddr := flag.String("proxy", "", "route via proxy (empty for no proxy), e.g. 127.0.0.1:9998")
	flag.Parse()
	endpoint := fmt.Sprintf("%s:%s", *listenAddr, *listenPort)

	golog.Infof("dialer starting [%s] ...", endpoint)

	var (
		err  error
		conn net.Conn
	)
	if *proxyAddr == "" {
		conn, err = net.Dial("tcp", endpoint)
		if err != nil {
			golog.Fatal(err)
		}
	} else {
		golog.Debug("connecting via proxy ...")
		conn, err = net.Dial("tcp", *proxyAddr)
		if err != nil {
			golog.Fatal(err)
		}

		_, err = conn.Write([]byte("client"))
		if err != nil {
			golog.Fatal(err)
		}
	}

	golog.Infof("connection established with %s", conn.RemoteAddr())

	if err := keyboard.Open(); err != nil {
		panic(err)
	}
	defer func() {
		if err := keyboard.Close(); err != nil {
			golog.Error(err)
		}
	}()

	golog.Warn("press ESC to quit")
	for {
		keyRune, key, err := keyboard.GetKey()
		if err != nil {
			golog.Fatal(err)
		}

		switch {
		case key == keyboard.KeyArrowUp:
			golog.Debugf("^")
			_, err = conn.Write([]byte(desktop_laptop_connection.ControlIncrease))
			if err != nil {
				golog.Error(err)
			}
		case key == keyboard.KeyArrowDown:
			golog.Debugf("v")
			_, err = conn.Write([]byte(desktop_laptop_connection.ControlDecrease))
			if err != nil {
				golog.Error(err)
			}
		case key == keyboard.KeyEsc,
			keyRune == 113: // keyRune 113 = q
			golog.Warn("closing connection")
			if err := conn.Close(); err != nil {
				golog.Error(err)
			}
			return
		}
	}
}
