package desktop_laptop_connection

import (
	"errors"
	"net"

	"github.com/kataras/golog"
)

const (
	ControlIncrease = "1"
	ControlDecrease = "2"
)

func GetLocalIp() (string, error) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, i := range netInterfaces {
		addrs, err := i.Addrs()
		if err != nil {
			golog.Error(err)
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}

			return ip.String(), nil
		}
	}

	return "", errors.New("seems like you are not connected to the network")
}
