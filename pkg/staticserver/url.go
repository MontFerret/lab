package staticserver

import "github.com/MontFerret/lab/v2/pkg/localserver"

func endpointURL(host string, port int) string {
	return localserver.EndpointURL(host, port)
}
