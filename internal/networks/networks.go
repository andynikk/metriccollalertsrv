package networks

import (
	"net"

	"github.com/andynikk/metriccollalertsrv/internal/constants"
)

func AddressAllowed(IPs []string, AllowedIP string) bool {
	_, ipv4Net, _ := net.ParseCIDR(AllowedIP)

	if ipv4Net == nil {
		return true
	}

	for _, sIP := range IPs {
		ip := net.ParseIP(sIP)

		if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return false
		}

		if ipv4Net.Contains(ip) {
			return true
		}
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
