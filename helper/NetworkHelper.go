package helper

import (
	"net"
	"os"
	"strings"
)

// GetOutboundAddress finds the best outbound address to best get around docker's network fuckery
// Which is ideally the IP address of the container, but if we can't get that the hostname might work
func GetOutboundAddress() string {
	ifaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range ifaces {
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				ipString := ip.String()
				// The address we're looking for begin with 10. so assume it's that address
				if strings.HasPrefix(ipString, "10.") {
					return ipString
				}
			}
		}
	}

	hostname, _ := os.Hostname()
	return hostname
}
