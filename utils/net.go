package utils

import (
	"net"
	"strings"
	"time"
)

// ExternalIP gets ip from given interface.
// If interface name is an IP then the passed IP is returned.
// If interface is empty then a connection attempt if performed
// to determine the default interface for external traffic.
func ExternalIP(ifaceName string) (net.IP, error) {
	if ifaceName != "" {
		if ip := net.ParseIP(ifaceName); ip != nil {
			return ip, nil
		}

		iface, err := net.InterfaceByName(ifaceName)
		if err != nil {
			return nil, err
		}
		return getIPFromInterface(iface)
	}

	// Connect to internet destination and check the local address
	c, err := net.DialTimeout("tcp", "api.open.wisdom.alphasoc.net:443", 5*time.Second)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	h, _, err := net.SplitHostPort(c.LocalAddr().String())
	if err != nil {
		return nil, err
	}
	return net.ParseIP(h), nil
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
		if ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			continue
		}
		return ip, nil
	}
	return nil, nil
}

func FQDN(h string) string {
	return strings.TrimRight(h, ".") + "."
}
