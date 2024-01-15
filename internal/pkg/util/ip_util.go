package util

import (
	"net"
)

func ParseCIDR(cidr string) ([]string, error) {
	var ipStrings []string
	// 解析CIDR地址
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	for ip = ip.Mask(ipNet.Mask); ipNet.Contains(ip); incIP(ip) {
		ipStrings = append(ipStrings, ip.To4().String())
	}
	return ipStrings, nil
}

// 增加IP地址函数
func incIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] > 0 {
			break
		}
	}
}
