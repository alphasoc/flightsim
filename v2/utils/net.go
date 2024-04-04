package utils

import (
	"fmt"
	"net"
	"strings"
	"time"
)

// UsableIP gets a "usable" IP address from the given interface.
// If ifaceName is an IP then the passed IP is returned.  If ifaceName is an interface,
// then the first external IP (if present) is returned, otherwise the first internal IP
// (if present) is returned, otherwise nil and an error are returned.
// If ifaceName is empty then a connection attempt is performed to determine the default
// interface for external traffic.
func UsableIP(ifaceName string) (net.IP, error) {
	if ifaceName != "" {
		// We have an IP, try to parse it.
		if ip := net.ParseIP(ifaceName); ip != nil {
			return ip, nil
		}
		// We have (hopefully) an interface.  Collect all addresses, separating external
		// from internal.
		iface, err := net.InterfaceByName(ifaceName)
		if err != nil {
			return nil, err
		}
		ips, err := getAddrsFromInterface(iface)
		if err != nil {
			return nil, err
		}
		var extIPs []net.IP
		var intIPs []net.IP
		for _, ip := range ips {
			if IsExternalIP(ip) {
				extIPs = append(extIPs, ip)
			} else {
				intIPs = append(intIPs, ip)
			}
		}
		// Return an external IP if possible, but fallback to an internal IP.
		if len(extIPs) > 0 {
			return extIPs[0], nil
		} else if len(intIPs) > 0 {
			return intIPs[0], nil
		}
		// No external, no internal... bad, bad interface.
		return nil, fmt.Errorf("no IP addresses found")
	}
	// ifaceName is empty, so try to grab a usable external IP by connecting to an internet
	// destination and checking the local address.
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

func getAddrsFromInterface(iface *net.Interface) ([]net.IP, error) {
	netAddrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}
	var ips []net.IP
	for _, addr := range netAddrs {
		switch t := addr.(type) {
		case *net.IPAddr:
			ips = append(ips, t.IP)
		case *net.IPNet:
			ips = append(ips, t.IP)
		}
	}
	return ips, nil
}

// IsExternalIP returns true if ip is an external IP (ie. not a loopback or link-local).
func IsExternalIP(ip net.IP) bool {
	return !(ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast())
}

func FQDN(h string) string {
	return strings.TrimRight(h, ".") + "."
}
