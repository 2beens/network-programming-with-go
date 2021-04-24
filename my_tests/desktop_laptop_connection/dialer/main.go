package main

import (
	"net"

	"github.com/2beens/network-programming-with-go/my_tests/desktop_laptop_connection"
	"github.com/eiannone/keyboard"
	"github.com/kataras/golog"
)

func main() {
	golog.SetLevel("debug")
	golog.Infof("dialer starting [%s] ...", desktop_laptop_connection.ListenerAddress)

	conn, err := net.Dial("tcp", desktop_laptop_connection.ListenerAddress)
	if err != nil {
		golog.Fatal(err)
	}

	golog.Infof("connection established")

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
		_, key, err := keyboard.GetKey()
		if err != nil {
			golog.Fatal(err)
		}

		switch key {
		case keyboard.KeyArrowUp:
			golog.Debugf("^")
			_, err = conn.Write([]byte(desktop_laptop_connection.ControlIncrease))
			if err != nil {
				golog.Error(err)
			}
		case keyboard.KeyArrowDown:
			golog.Debugf("v")
			_, err = conn.Write([]byte(desktop_laptop_connection.ControlDecrease))
			if err != nil {
				golog.Error(err)
			}
		case keyboard.KeyEsc:
			golog.Warn("closing connection")
			if err := conn.Close(); err != nil {
				golog.Error(err)
			}
			return
		}
	}
}
