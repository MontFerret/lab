package staticserver

import "github.com/MontFerret/lab/v2/pkg/localserver"

func GetFreePort(host string) (int, error) {
	return localserver.GetFreePort(host)
}
