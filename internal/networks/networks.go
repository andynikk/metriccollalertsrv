package networks

import (
	"net"

	"github.com/andynikk/metriccollalertsrv/internal/constants"
)

func AddressAllowed(IPAgent string, ipv4Net *net.IPNet) bool {

	ip := net.ParseIP(IPAgent)

	if ipv4Net.Contains(ip) {
		return true
	}

	return false
}

func IPv4RangesToStr(IPs []net.IP) string {

	strIP := ""
	for _, ip := range IPs {
		if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			continue
		}

		strIP += ip.String() + constants.SepIPAddress
	}

	return strIP
}
