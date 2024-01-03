package utils

import (
	"net"
	"strings"
)

func IsDomain(host string) bool {
	if host == "" {
		return false
	}

	hostList := strings.Split(host, ":")

	return net.ParseIP(hostList[0]) == nil
}
