package utils

import (
	"errors"
	"net"
)

// ExternalIP gets ip from given interface or
// if interface is empty it finds first public ip.
func ExternalIP(ifaceName string) (net.IP, error) {
	if ifaceName != "" {
		iface, err := net.InterfaceByName(ifaceName)
		if err != nil {
			return nil, err
		}
		return getIPFromInterface(iface)
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 ||
			iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		ip, err := getIPFromInterface(&iface)
		if err != nil {
			return nil, err
		}
		if ip != nil {
			return ip, nil
		}

	}
	return nil, errors.New("local interfaces have no public ip")
}

func getIPFromInterface(iface *net.Interface) (net.IP, error) {
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPAddr:
			ip = v.IP
		case *net.IPNet:
			ip = v.IP
		}
		if ip == nil || ip.IsLoopback() {
			continue
		}
		return ip, nil
	}
	return nil, nil
}
