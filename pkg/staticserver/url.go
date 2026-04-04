package staticserver

import (
	"net"
	"strconv"
)

func endpointURL(host string, port int) string {
	return "http://" + net.JoinHostPort(host, strconv.Itoa(port))
}
