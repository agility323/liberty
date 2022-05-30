/*
By Thomas Wade, 2022.05.30
*/
package lbtutil

import (
	"net"
	"strings"
	"strconv"
)

func GetLocalIP() (string, error) {
	itfs, err := net.Interfaces()
	if err != nil { return "", err }
	for _, itf := range itfs {
		if itf.Flags & net.FlagUp == 0 { continue }
		if itf.Flags & net.FlagLoopback != 0 { continue }
		addrs, err := itf.Addrs()
		if err != nil { return "", err }
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			default:
				continue
			}
			if ip.IsLoopback() { continue }
			if ip = ip.To4(); ip == nil { continue }
			// local ip only: 10.0.0.0/8 192.168.0.0/16 172.16.0.0/12
			ipstr := ip.String()
			if strings.HasPrefix(ipstr, "10.") || strings.HasPrefix(ipstr, "192.168.") { return ipstr, nil }
			if strings.HasPrefix(ipstr, "172.") {
				segs := strings.SplitN(ipstr, ":", 3)
				if len(segs) == 3 {
					if n, err := strconv.Atoi(segs[1]); err == nil && n >= 16 && n < 32 { return ipstr, nil }
				}
			}
		}
	}
	return "", nil
}
