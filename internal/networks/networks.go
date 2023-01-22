package networks

import (
	"net"
)

func AddressAllowed(IPAgent string, ipv4Net *net.IPNet) bool {
	ip := net.ParseIP(IPAgent)
	return ipv4Net.Contains(ip)
}
