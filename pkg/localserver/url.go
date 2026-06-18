package localserver

import (
	"net"
	"strconv"
)

func EndpointURL(host string, port int) string {
	return "http://" + net.JoinHostPort(host, strconv.Itoa(port))
}
